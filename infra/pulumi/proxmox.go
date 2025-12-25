package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"

	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve/download"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve/storage"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve/vm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Cluster config
type ClusterConfig struct {
	Username string `koanf:"username"`
	Endpoint string `koanf:"endpoint"`
	NodeName string `koanf:"nodename"`
	ImageUrl string `koanf:"image_url"`
}

// Node config
type NodeConfig struct {
	Name     string
	Role     string
	Cores    int
	Memory   int
	DiskSize int
	HasGPU   bool
	PcieIDs  []string
}

func loadConfig() (*ClusterConfig, string, error) {
	if err := k.Load(file.Provider("config.yml"), yaml.Parser()); err != nil {
		return nil, "", fmt.Errorf("error loading config.yml: %v", err)
	}

	var clusterConfig ClusterConfig
	if err := k.Unmarshal("proxmox", &clusterConfig); err != nil {
		return nil, "", fmt.Errorf("error unmarshalling proxmox config: %v", err)
	}

	password := k.String("PROXMOX_PASSWORD")
	return &clusterConfig, password, nil
}

func initializeProvider(ctx *pulumi.Context, cfg *ClusterConfig, password string) (*proxmoxve.Provider, error) {
	if cfg.Username == "" || cfg.Endpoint == "" {
		return nil, fmt.Errorf("set username and endpoint in config.yml")
	}
	if password == "" {
		return nil, fmt.Errorf("PROXMOX_PASSWORD not set in .env")
	}

	return proxmoxve.NewProvider(ctx, "proxmoxve", &proxmoxve.ProviderArgs{
		Endpoint: pulumi.String(cfg.Endpoint),
		Username: pulumi.String(cfg.Username),
		Password: pulumi.String(password),
		Insecure: pulumi.Bool(true),
	})
}

func createCloudInit(ctx *pulumi.Context, provider *proxmoxve.Provider, nodeConfig NodeConfig, clusterNodeName string) (*storage.File, error) {
	data, err := os.ReadFile(filepath.Join("cloud-init", "cloud-init.yml"))
	if err != nil {
		return nil, err
	}

	content := strings.ReplaceAll(string(data), "${hostname}", nodeConfig.Name)
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
		Cpu: &vm.VirtualMachineCpuArgs{
			Cores: pulumi.Int(config.Cores),
			Type:  pulumi.String("host"),
		},
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
		Bios:    pulumi.String("ovmf"),
		Machine: pulumi.String("q35"),
		EfiDisk: &vm.VirtualMachineEfiDiskArgs{
			DatastoreId: pulumi.String("local-lvm"),
			Type:        pulumi.String("4m"),
		},
	}
	// Set GPU specific configurations
	if config.HasGPU {
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

func DeployProxmox(ctx *pulumi.Context) error {
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
	// Download cloud image to Proxmox
	cloudImage, err := download.NewFile(ctx, "ubuntu-cloud-image", &download.FileArgs{
		ContentType: pulumi.String("iso"),
		DatastoreId: pulumi.String("local"),
		NodeName:    pulumi.String(clusterConfig.NodeName),
		Url:         pulumi.String(clusterConfig.ImageUrl),
		FileName:    pulumi.String("ubuntu-main.img"),
		Overwrite:   pulumi.Bool(true),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}
	// Define cluster nodes
	nodes := []NodeConfig{
		{Name: "k8s-controller1", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30},
		{Name: "k8s-controller2", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30},
		{Name: "k8s-controller3", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30},
		{Name: "k8s-worker1", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125},
		{Name: "k8s-worker2", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125},
		{Name: "k8s-worker3", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125},
		{Name: "k8s-worker4", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125, HasGPU: true, PcieIDs: []string{"0000:28:00.0"}},
	}
	// Create VMs for each node
	for _, node := range nodes {
		// Create cloud-init config
		cloudInit, err := createCloudInit(ctx, provider, node, clusterConfig.NodeName)
		if err != nil {
			return err
		}
		// Create VM
		vm, err := createVM(ctx, provider, clusterConfig.NodeName, node, cloudInit, cloudImage)
		if err != nil {
			// TODO: remove it later
			// return err
			continue
		}
		// Export VM details
		ctx.Export(node.Name, pulumi.Map{
			"id": vm.ID(),
			"ip": vm.Ipv4Addresses,
		})
	}
	return nil
}
