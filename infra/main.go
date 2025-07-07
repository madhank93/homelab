package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/muhlba91/pulumi-proxmoxve/sdk/v6/go/proxmoxve"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v6/go/proxmoxve/download"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v6/go/proxmoxve/storage"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v6/go/proxmoxve/vm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Cluster config structs
type ClusterConfig struct {
	Username string `koanf:"username"`
	Endpoint string `koanf:"endpoint"`
	NodeName string `koanf:"nodename"`
	ImageUrl string `koanf:"image_url"`
}

// NodeConfig remains in code
type NodeConfig struct {
	Name     string
	Role     string
	Cores    int
	Memory   int
	DiskSize int
	HasGPU   bool
	PcieIDs  []string
}

var k = koanf.New(".")

func loadConfig() (*ClusterConfig, string, error) {
	// Load YAML config
	if err := k.Load(file.Provider("config.yml"), yaml.Parser()); err != nil {
		return nil, "", fmt.Errorf("error loading config.yml: %v", err)
	}
	// Load .env for sensitive env vars (password)
	if err := k.Load(file.Provider(".env"), dotenv.Parser()); err != nil {
		return nil, "", fmt.Errorf("error loading .env: %v", err)
	}
	// Unmarshal cluster config
	var clusterConfig ClusterConfig
	if err := k.Unmarshal("proxmox", &clusterConfig); err != nil {
		return nil, "", fmt.Errorf("error un-marshalling proxmox config: %v", err)
	}
	password := k.String("PROXMOX_PASSWORD")
	return &clusterConfig, password, nil
}

func initializeProvider(ctx *pulumi.Context, cfg *ClusterConfig, password string) (*proxmoxve.Provider, error) {
	if cfg.Username == "" || cfg.Endpoint == "" {
		return nil, fmt.Errorf("please set username and endpoint in the config.yml file")
	}
	if password == "" {
		return nil, fmt.Errorf("PROXMOX_PASSWORD must be set in the .env file")
	}
	return proxmoxve.NewProvider(ctx, "proxmoxve", &proxmoxve.ProviderArgs{
		Endpoint: pulumi.String(cfg.Endpoint),
		Username: pulumi.String(cfg.Username),
		Password: pulumi.String(password),
		Insecure: pulumi.Bool(true),
	})
}

func createCloudInit(ctx *pulumi.Context, provider *proxmoxve.Provider, nodeConfig NodeConfig, clusterNodeName string) (*storage.File, error) {
	configPath := filepath.Join("cloud-init", "cloud-init.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	content := string(data)
	replacements := map[string]string{
		"${hostname}":          nodeConfig.Name,
		"${network_interface}": "ens18",
	}
	if nodeConfig.HasGPU {
		replacements["${network_interface}"] = "enp6s18"
	}
	for k, v := range replacements {
		content = strings.ReplaceAll(content, k, v)
	}
	return storage.NewFile(ctx, nodeConfig.Name+"-cloud-init", &storage.FileArgs{
		NodeName:    pulumi.String(clusterNodeName),
		DatastoreId: pulumi.String("local"),
		ContentType: pulumi.String("snippets"),
		FileMode:    pulumi.String("0755"),
		Overwrite:   pulumi.Bool(true),
		SourceRaw: storage.FileSourceRawArgs{
			Data:     pulumi.String(content),
			FileName: pulumi.String(nodeConfig.Name + ".yml"),
		},
	}, pulumi.Provider(provider))
}

func createVM(ctx *pulumi.Context, provider *proxmoxve.Provider, nodeName string, config NodeConfig, cloudInit *storage.File, cloudImage *download.File) (*vm.VirtualMachine, error) {
	args := &vm.VirtualMachineArgs{
		NodeName: pulumi.String(nodeName),
		Name:     pulumi.String(config.Name),
		Agent:    vm.VirtualMachineAgentArgs{Enabled: pulumi.Bool(true)},
		Cpu:      &vm.VirtualMachineCpuArgs{Cores: pulumi.Int(config.Cores)},
		Memory: &vm.VirtualMachineMemoryArgs{
			Dedicated: pulumi.Int(config.Memory),
			Floating:  pulumi.Int(0),
		},
		Disks: &vm.VirtualMachineDiskArray{
			vm.VirtualMachineDiskArgs{
				Size:       pulumi.Int(config.DiskSize),
				Interface:  pulumi.String("scsi0"),
				Iothread:   pulumi.Bool(true),
				FileFormat: pulumi.String("raw"),
				FileId:     cloudImage.ID(),
			},
		},
		BootOrders:   pulumi.StringArray{pulumi.String("scsi0"), pulumi.String("net0")},
		ScsiHardware: pulumi.String("virtio-scsi-single"),
		OperatingSystem: vm.VirtualMachineOperatingSystemArgs{
			Type: pulumi.String("l26"),
		},
		NetworkDevices: vm.VirtualMachineNetworkDeviceArray{
			vm.VirtualMachineNetworkDeviceArgs{
				Model:  pulumi.String("virtio"),
				Bridge: pulumi.String("vmbr0"),
			},
		},
		OnBoot: pulumi.Bool(true),
		Initialization: &vm.VirtualMachineInitializationArgs{
			DatastoreId:    pulumi.String("local-lvm"),
			UserDataFileId: cloudInit.ID(),
		},
	}
	// Set GPU specific configurations
	if config.HasGPU {
		args.Bios = pulumi.String("ovmf")
		args.Machine = pulumi.String("q35")
		var hostpcis vm.VirtualMachineHostpciArray
		for i, pcieID := range config.PcieIDs {
			hostpcis = append(hostpcis, vm.VirtualMachineHostpciArgs{
				Device: pulumi.String("hostpci0"),
				Pcie:   pulumi.Bool(true),
				Id:     pulumi.String(pcieID),
				Xvga:   pulumi.Bool(i == 0),
			})
		}
		args.Hostpcis = hostpcis
	}
	return vm.NewVirtualMachine(ctx, config.Name, args,
		pulumi.DependsOn([]pulumi.Resource{cloudImage}),
		pulumi.Provider(provider))
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Load config and secrets
		clusterConfig, password, err := loadConfig()
		if err != nil {
			return err
		}
		// Initialize Proxmox provider
		provider, err := initializeProvider(ctx, clusterConfig, password)
		if err != nil {
			return err
		}
		// Download cloud image
		cloudImage, err := download.NewFile(ctx, "ubuntu-cloud-image", &download.FileArgs{
			ContentType: pulumi.String("iso"),
			DatastoreId: pulumi.String("local"),
			NodeName:    pulumi.String(clusterConfig.NodeName),
			Url:         pulumi.String(clusterConfig.ImageUrl),
			Overwrite:   pulumi.Bool(true),
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}
		// Define node configurations
		nodes := []NodeConfig{
			{Name: "k8s-controller1", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30, HasGPU: false},
			{Name: "k8s-controller2", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30, HasGPU: false},
			{Name: "k8s-controller3", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30, HasGPU: false},
			{Name: "k8s-worker1", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125, HasGPU: false},
			{Name: "k8s-worker2", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125, HasGPU: false},
			{Name: "k8s-worker3", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125, HasGPU: false},
			{Name: "k8s-worker4", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125, HasGPU: true, PcieIDs: []string{"0000:28:00.0"}},
		}
		// Create VMs
		for _, node := range nodes {
			// Prepare cloud-init configs and upload
			cloudInit, err := createCloudInit(ctx, provider, node, clusterConfig.NodeName)
			if err != nil {
				return err
			}
			// Create VMs
			vm, err := createVM(ctx, provider, clusterConfig.NodeName, node, cloudInit, cloudImage)
			if err != nil {
				return err
			}
			// Export VM ID and IP address
			ctx.Export(node.Name, pulumi.Map{
				"id": vm.ID(),
				"ip": vm.Ipv4Addresses,
			})
		}
		return nil
	})
}
