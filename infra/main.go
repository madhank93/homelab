package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v6/go/proxmoxve"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v6/go/proxmoxve/download"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v6/go/proxmoxve/storage"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v6/go/proxmoxve/vm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Config structs
type ClusterConfig struct {
	NodeName    string
	DatastoreID string
	Bridge      string
	ImageUrl    string
	Nodes       []NodeConfig
}

type NodeConfig struct {
	Name     string
	Role     string
	Cores    int
	Memory   int
	DiskSize int
	HasGPU   bool
}

// initialize the Proxmox provider
func initializeProvider(ctx *pulumi.Context) (*proxmoxve.Provider, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	username := os.Getenv("PROXMOX_USERNAME")
	password := os.Getenv("PROXMOX_PASSWORD")
	endpoint := os.Getenv("PROXMOX_ENDPOINT")

	if username == "" || password == "" {
		return nil, fmt.Errorf("PROXMOX_USERNAME and PROXMOX_PASSWORD must be set in the .env file")
	}

	if endpoint == "" {
		endpoint = "https://192.168.1.198:8006/"
	}

	return proxmoxve.NewProvider(ctx, "proxmoxve", &proxmoxve.ProviderArgs{
		Endpoint: pulumi.String(endpoint),
		Username: pulumi.String(username),
		Password: pulumi.String(password),
		Insecure: pulumi.Bool(true),
	})
}

// generates and uploads cloud-init config
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

// creates a VM
func createVM(ctx *pulumi.Context, provider *proxmoxve.Provider, nodeName string, config NodeConfig, cloudInit *storage.File, cloudImage *download.File, bridge string) (*vm.VirtualMachine, error) {
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
				Bridge: pulumi.String(bridge),
			},
		},
		OnBoot: pulumi.Bool(true),
		Initialization: &vm.VirtualMachineInitializationArgs{
			DatastoreId:    pulumi.String("local-lvm"),
			UserDataFileId: cloudInit.ID(),
		},
	}

	if config.HasGPU {
		args.Bios = pulumi.String("ovmf")
		args.Machine = pulumi.String("q35")
		args.Hostpcis = vm.VirtualMachineHostpciArray{
			vm.VirtualMachineHostpciArgs{
				Device: pulumi.String("hostpci0"),
				Pcie:   pulumi.Bool(true),
				Id:     pulumi.String("0000:17:00.0"),
			},
		}
	}

	return vm.NewVirtualMachine(ctx, config.Name, args,
		pulumi.DependsOn([]pulumi.Resource{cloudImage}),
		pulumi.Provider(provider))
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Initialize provider
		provider, err := initializeProvider(ctx)
		if err != nil {
			return err
		}

		// Define cluster configs
		clusterConfig := ClusterConfig{
			NodeName:    "pve",
			DatastoreID: "local",
			Bridge:      "vmbr0",
			ImageUrl:    "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img",
			Nodes: []NodeConfig{
				{Name: "k8s-controller1", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30, HasGPU: false},
				{Name: "k8s-controller2", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30, HasGPU: false},
				{Name: "k8s-controller3", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30, HasGPU: false},
				{Name: "k8s-worker1", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125, HasGPU: false},
				{Name: "k8s-worker2", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125, HasGPU: false},
				{Name: "k8s-worker3", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125, HasGPU: false},
				{Name: "k8s-worker4", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125, HasGPU: false},
			},
		}

		// Download cloud image
		cloudImage, err := download.NewFile(ctx, "ubuntu-cloud-image", &download.FileArgs{
			ContentType: pulumi.String("iso"),
			DatastoreId: pulumi.String(clusterConfig.DatastoreID),
			NodeName:    pulumi.String(clusterConfig.NodeName),
			Url:         pulumi.String(clusterConfig.ImageUrl),
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		for _, node := range clusterConfig.Nodes {
			// Prepare cloud-init configs and upload
			cloudInit, err := createCloudInit(ctx, provider, node, clusterConfig.NodeName)
			if err != nil {
				return err
			}

			// Create VMs
			vm, err := createVM(ctx, provider, clusterConfig.NodeName, node, cloudInit, cloudImage, clusterConfig.Bridge)
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
