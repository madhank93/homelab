package platform

import (
	"fmt"

	"github.com/muhlba91/pulumi-proxmoxve/sdk/v8/go/proxmoxve"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v8/go/proxmoxve/download"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/core/internal/cfg"
)

// NewProxmoxProvider initialises the Proxmox VE provider from config.yml and
// the PROXMOX_PASSWORD environment variable (injected via SOPS).
func NewProxmoxProvider(ctx *pulumi.Context) (*proxmoxve.Provider, error) {
	pcfg := config.C.Proxmox
	password := cfg.K.String("PROXMOX_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("PROXMOX_PASSWORD not found in .env")
	}

	provider, err := proxmoxve.NewProvider(ctx, "proxmoxve", &proxmoxve.ProviderArgs{
		Endpoint: pulumi.String(pcfg.Endpoint),
		Username: pulumi.String(pcfg.Username),
		Password: pulumi.String(password),
		Insecure: pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return provider, nil
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
func DownloadImage(ctx *pulumi.Context, provider *proxmoxve.Provider, resourceName, nodeName, url, fileName, compression string) (*download.FileLegacy, error) {
	return download.NewFileLegacy(ctx, resourceName, &download.FileLegacyArgs{
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
// Each of the node's PcieIDs becomes a PCIe passthrough device with Xvga=false
// (compute-only; no VGA BAR constraints).
func NewProxmoxVM(ctx *pulumi.Context, provider *proxmoxve.Provider, nodeName string, node config.Node, imageID pulumi.IDInput) (*proxmoxve.VmLegacy, error) {
	diskArgs := proxmoxve.VmLegacyDiskArgs{
		Size:        pulumi.Int(node.Disk),
		Interface:   pulumi.String("scsi0"),
		Iothread:    pulumi.Bool(true),
		FileFormat:  pulumi.String("raw"),
		FileId:      imageID.ToIDOutput(),
		DatastoreId: pulumi.String("local-lvm"),
	}

	args := &proxmoxve.VmLegacyArgs{
		NodeName: pulumi.String(nodeName),
		Name:     pulumi.String(node.Name),
		Agent:    proxmoxve.VmLegacyAgentArgs{Enabled: pulumi.Bool(true)},
		Cpu: &proxmoxve.VmLegacyCpuArgs{
			Cores: pulumi.Int(node.Cores),
			Type:  pulumi.String("host"),
			Units: pulumi.Int(node.CPUUnits), // CPU Weight
		},
		Memory: &proxmoxve.VmLegacyMemoryArgs{
			Dedicated: pulumi.Int(node.Memory),
			Shared:    pulumi.Int(node.Balloon), // Balloon Minimum (Guaranteed)
		},
		Disks: &proxmoxve.VmLegacyDiskArray{
			diskArgs,
		},
		BootOrders:   pulumi.StringArray{pulumi.String("scsi0")},
		ScsiHardware: pulumi.String("virtio-scsi-single"),
		OperatingSystem: proxmoxve.VmLegacyOperatingSystemArgs{
			Type: pulumi.String("l26"), // Linux 2.6+
		},
		NetworkDevices: proxmoxve.VmLegacyNetworkDeviceArray{
			proxmoxve.VmLegacyNetworkDeviceArgs{
				Model:  pulumi.String("virtio"),
				Bridge: pulumi.String("vmbr0"),
			},
		},
		OnBoot: pulumi.Bool(true),
		// Nodes reboot only through talosctl, one at a time; a config update must
		// never restart every VM at once.
		RebootAfterUpdate: pulumi.Bool(false),
		Bios:              pulumi.String("ovmf"),
		Machine:           pulumi.String("q35"),
		EfiDisk: &proxmoxve.VmLegacyEfiDiskArgs{
			DatastoreId: pulumi.String("local-lvm"),
			FileFormat:  pulumi.String("raw"),
			Type:        pulumi.String("4m"),
		},
	}

	if node.HasGPU() {
		var hostpcis proxmoxve.VmLegacyHostpciArray
		for i, pcieID := range node.PcieIDs {
			hostpcis = append(hostpcis, proxmoxve.VmLegacyHostpciArgs{
				Device: pulumi.String(fmt.Sprintf("hostpci%d", i)),
				Pcie:   pulumi.Bool(true),
				Id:     pulumi.String(pcieID),
				// Xvga=false: this is a compute-only GPU (CUDA/Ollama/ComfyUI), not a display.
				// xvga=true adds VGA BAR constraints that conflict with large memory allocations.
				Xvga: pulumi.Bool(false),
			})
		}
		args.Hostpcis = hostpcis
	}

	// Prevent accidental VM replacement; controllers are etcd members.
	// Unprotect explicitly for intentional rebuilds.
	return proxmoxve.NewVmLegacy(ctx, node.Name, args,
		pulumi.Provider(provider),
		pulumi.Protect(true),
		// The image only seeds a new VM's disk; a newer image must not replace
		// running nodes, which upgrade in place via `just talos-upgrade`.
		pulumi.IgnoreChanges([]string{"disks[0].fileId"}),
	)
}
