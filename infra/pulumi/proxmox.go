package main

import (
	"fmt"

	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve/download"
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
	Name       string
	IP         string
	Role       string
	Cores      int
	Memory     int
	DiskSize   int
	HasGPU     bool
	PcieIDs    []string
	IsTemplate bool
	TemplateID pulumi.IntInput
	CpuUnits   int
	Balloon    int
}

// Initializes the Proxmox provider
func NewProxmoxProvider(ctx *pulumi.Context) (*proxmoxve.Provider, *ProxmoxClusterConfig, error) {
	var cfg ProxmoxClusterConfig
	if err := LoadConfig("proxmox", &cfg); err != nil {
		return nil, nil, err
	}

	password := k.String("PROXMOX_PASSWORD")
	if password == "" {
		return nil, nil, fmt.Errorf("PROXMOX_PASSWORD not found in .env")
	}

	provider, err := proxmoxve.NewProvider(ctx, "proxmoxve", &proxmoxve.ProviderArgs{
		Endpoint: pulumi.String(cfg.Endpoint),
		Username: pulumi.String(cfg.Username),
		Password: pulumi.String(password),
		Insecure: pulumi.Bool(true),
	})
	if err != nil {
		return nil, nil, err
	}

	return provider, &cfg, nil
}

// DownloadImage downloads a file to the Proxmox datastore.
//
// Reliability notes:
//   - IgnoreChanges("fileName"): prevents state drift from triggering unintended
//     replacements. If the resource is dropped from state (e.g. after pulumi cancel),
//     Pulumi will reconcile without forcing a re-download.
//   - DeleteBeforeReplace: when a replacement IS intentional (URL/schematic change),
//     Pulumi deletes the Proxmox file first (physically removing it from disk), then
//     re-downloads — eliminating the "refusing to override existing file" error.
//   - OverwriteUnmanaged: safety net for files that exist on Proxmox but were never
//     in Pulumi state (e.g. manually uploaded files with the same name).
func DownloadImage(ctx *pulumi.Context, provider *proxmoxve.Provider, resourceName, nodeName, url, fileName, compression string) (*download.File, error) {
	return download.NewFile(ctx, resourceName, &download.FileArgs{
		ContentType:            pulumi.String("iso"),
		DatastoreId:            pulumi.String("local"),
		NodeName:               pulumi.String(nodeName),
		Url:                    pulumi.String(url),
		FileName:               pulumi.String(fileName),
		DecompressionAlgorithm: pulumi.String(compression),
		Overwrite:              pulumi.Bool(true),
		OverwriteUnmanaged:     pulumi.Bool(true),
	},
		pulumi.Provider(provider),
		pulumi.IgnoreChanges([]string{"fileName"}),
		pulumi.DeleteBeforeReplace(true),
	)
}

// Virtual Machine
func NewProxmoxVM(ctx *pulumi.Context, provider *proxmoxve.Provider, nodeName string, config NodeConfig, imageID pulumi.IDInput) (*vm.VirtualMachine, error) {
	diskArgs := vm.VirtualMachineDiskArgs{
		Size:        pulumi.Int(config.DiskSize),
		Interface:   pulumi.String("scsi0"),
		Iothread:    pulumi.Bool(true),
		FileFormat:  pulumi.String("raw"),
		FileId:      imageID.ToIDOutput(),
		DatastoreId: pulumi.String("local-lvm"),
	}

	args := &vm.VirtualMachineArgs{
		NodeName: pulumi.String(nodeName),
		Name:     pulumi.String(config.Name),
		Agent:    vm.VirtualMachineAgentArgs{Enabled: pulumi.Bool(true)},
		Cpu: &vm.VirtualMachineCpuArgs{
			Cores: pulumi.Int(config.Cores),
			Type:  pulumi.String("host"),
			Units: pulumi.Int(config.CpuUnits), // CPU Weight
		},
		Memory: &vm.VirtualMachineMemoryArgs{
			Dedicated: pulumi.Int(config.Memory),
			Shared:    pulumi.Int(config.Balloon), // Balloon Minimum (Guaranteed)
		},
		Disks: &vm.VirtualMachineDiskArray{
			diskArgs,
		},
		BootOrders:   pulumi.StringArray{pulumi.String("scsi0")},
		ScsiHardware: pulumi.String("virtio-scsi-single"),
		OperatingSystem: vm.VirtualMachineOperatingSystemArgs{
			Type: pulumi.String("l26"), // Linux 2.6+
		},
		NetworkDevices: vm.VirtualMachineNetworkDeviceArray{
			vm.VirtualMachineNetworkDeviceArgs{
				Model:  pulumi.String("virtio"),
				Bridge: pulumi.String("vmbr0"),
			},
		},
		OnBoot:  pulumi.Bool(true),
		Bios:    pulumi.String("ovmf"),
		Machine: pulumi.String("q35"),
		EfiDisk: &vm.VirtualMachineEfiDiskArgs{
			DatastoreId: pulumi.String("local-lvm"),
			FileFormat:  pulumi.String("raw"),
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

	return vm.NewVirtualMachine(ctx, config.Name, args, pulumi.Provider(provider))
}
