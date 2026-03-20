package platform

import (
	"fmt"

	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve/download"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve/vm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/madhank93/homelab/core/internal/cfg"
)

// ProxmoxClusterConfig holds Proxmox provider settings read from config.yml
// under the "proxmox" key.
type ProxmoxClusterConfig struct {
	Username string `koanf:"username"`
	Endpoint string `koanf:"endpoint"`
	NodeName string `koanf:"nodename"`
	ImageURL string `koanf:"image_url"`
}

// NodeConfig describes a single Talos VM to create on Proxmox.
// HasGPU and PcieIDs enable PCIe passthrough for the GPU worker node.
// CpuUnits controls the Proxmox CPU weight (scheduling priority).
// Balloon is the minimum guaranteed RAM in MiB (balloon floor).
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

// NewProxmoxProvider initialises the Proxmox VE provider from config.yml and
// the PROXMOX_PASSWORD environment variable (injected via SOPS).
func NewProxmoxProvider(ctx *pulumi.Context) (*proxmoxve.Provider, *ProxmoxClusterConfig, error) {
	var pcfg ProxmoxClusterConfig
	if err := cfg.Load("proxmox", &pcfg); err != nil {
		return nil, nil, err
	}

	password := cfg.K.String("PROXMOX_PASSWORD")
	if password == "" {
		return nil, nil, fmt.Errorf("PROXMOX_PASSWORD not found in .env")
	}

	provider, err := proxmoxve.NewProvider(ctx, "proxmoxve", &proxmoxve.ProviderArgs{
		Endpoint: pulumi.String(pcfg.Endpoint),
		Username: pulumi.String(pcfg.Username),
		Password: pulumi.String(password),
		Insecure: pulumi.Bool(true),
	})
	if err != nil {
		return nil, nil, err
	}

	return provider, &pcfg, nil
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

// NewProxmoxVM creates a Talos virtual machine on the given Proxmox node.
// It uses OVMF BIOS + q35 machine type, virtio-scsi-single, and an EFI disk.
// When config.HasGPU is true, each PcieID is added as a PCIe passthrough device
// with Xvga=false (compute-only; no VGA BAR constraints).
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
		for _, pcieID := range config.PcieIDs {
			hostpcis = append(hostpcis, vm.VirtualMachineHostpciArgs{
				Device: pulumi.String("hostpci0"),
				Pcie:   pulumi.Bool(true),
				Id:     pulumi.String(pcieID),
				// Xvga=false: this is a compute-only GPU (CUDA/Ollama/ComfyUI), not a display.
				// xvga=true adds VGA BAR constraints that conflict with large memory allocations.
				Xvga: pulumi.Bool(false),
			})
		}
		args.Hostpcis = hostpcis
	}

	return vm.NewVirtualMachine(ctx, config.Name, args, pulumi.Provider(provider))
}
