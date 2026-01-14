package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve/download"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve/storage"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve/vm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Proxmox-specific config structs
type ProxmoxClusterConfig struct {
	Username string `koanf:"username"`
	Endpoint string `koanf:"endpoint"`
	NodeName string `koanf:"nodename"`
	ImageUrl string `koanf:"image_url"`
}

type NodeConfig struct {
	Name     string
	Role     string
	Cores    int
	Memory   int
	DiskSize int
	HasGPU   bool
	PcieIDs  []string
}

func DeployProxmox(ctx *pulumi.Context) error {
	// 1. Generic Load
	var cfg ProxmoxClusterConfig
	if err := LoadConfig("proxmox", &cfg); err != nil {
		return err
	}

	// 2. Access Password from global k
	password := k.String("PROXMOX_PASSWORD")
	if password == "" {
		return fmt.Errorf("PROXMOX_PASSWORD not found in .env")
	}

	// Initialize Provider
	provider, err := proxmoxve.NewProvider(ctx, "proxmoxve", &proxmoxve.ProviderArgs{
		Endpoint: pulumi.String(cfg.Endpoint),
		Username: pulumi.String(cfg.Username),
		Password: pulumi.String(password),
		Insecure: pulumi.Bool(true),
	})
	if err != nil {
		return err
	}

	// Download Image
	cloudImage, err := download.NewFile(ctx, "ubuntu-cloud-image", &download.FileArgs{
		ContentType: pulumi.String("iso"),
		DatastoreId: pulumi.String("local"),
		NodeName:    pulumi.String(cfg.NodeName),
		Url:         pulumi.String(cfg.ImageUrl),
		FileName:    pulumi.String("ubuntu-main.img"),
		Overwrite:   pulumi.Bool(true),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	// Define Nodes
	nodes := []NodeConfig{
		{Name: "k8s-controller1", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30},
		{Name: "k8s-controller2", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30},
		{Name: "k8s-controller3", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30},
		{Name: "k8s-worker1", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125},
		{Name: "k8s-worker2", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125},
		{Name: "k8s-worker3", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125},
		{Name: "k8s-worker4", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125, HasGPU: true, PcieIDs: []string{"0000:28:00.0"}},
	}

	for _, node := range nodes {
		cloudInit, err := createCloudInit(ctx, provider, node, cfg.NodeName)
		if err != nil {
			return err
		}

		vmRes, err := createVM(ctx, provider, cfg.NodeName, node, cloudInit, cloudImage)
		if err != nil {
			fmt.Printf("Error creating VM %s: %v\n", node.Name, err)
			continue
		}

		ctx.Export(node.Name, pulumi.Map{
			"id": vmRes.ID(),
			"ip": vmRes.Ipv4Addresses,
		})
	}
	return nil
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
