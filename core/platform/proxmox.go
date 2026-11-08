package platform

import (
	"fmt"

	"github.com/muhlba91/pulumi-proxmoxve/sdk/v8/go/proxmoxve"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v8/go/proxmoxve/acme"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v8/go/proxmoxve/acme/dns"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v8/go/proxmoxve/download"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v8/go/proxmoxve/realm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/core/internal/cfg"
)

// NewProxmoxProvider returns the Proxmox VE provider, configured from config.yml
// and PROXMOX_PASSWORD.
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

// NewProxmoxCert orders a Let's Encrypt certificate for <proxmox.host>.<domain>
// through a Cloudflare DNS-01 challenge, so the web UI needs no internet exposure.
// Proxmox renews it itself (pve-daily-update) once it is within 30 days of expiry.
func NewProxmoxCert(ctx *pulumi.Context, provider *proxmoxve.Provider) error {
	pcfg := config.C.Proxmox
	cfToken := cfg.K.String("CLOUDFLARE_API_TOKEN")
	if cfToken == "" {
		return fmt.Errorf("CLOUDFLARE_API_TOKEN not found in .env")
	}

	account, err := acme.NewAccount(ctx, "pve-acme-account", &acme.AccountArgs{
		Name:      pulumi.String("letsencrypt"),
		Contact:   pulumi.String(config.C.ACMEEmail),
		Directory: pulumi.String("https://acme-v02.api.letsencrypt.org/directory"),
		Tos:       pulumi.String("https://letsencrypt.org/documents/LE-SA-v1.8-July-06-2026.pdf"),
	}, pulumi.Provider(provider), pulumi.IgnoreChanges([]string{"tos"}))
	if err != nil {
		return err
	}

	plugin, err := dns.NewPlugin(ctx, "pve-acme-cloudflare", &dns.PluginArgs{
		Plugin: pulumi.String("cloudflare"),
		Api:    pulumi.String("cf"),
		Data: pulumi.ToSecret(pulumi.StringMap{
			"CF_Token":   pulumi.String(cfToken),
			"CF_Zone_ID": pulumi.String(config.C.Cloudflare.ZoneID),
		}).(pulumi.StringMapOutput),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	_, err = acme.NewCertificate(ctx, "pve-acme-cert", &acme.CertificateArgs{
		NodeName: pulumi.String(pcfg.NodeName),
		Account:  account.Name,
		Domains: acme.CertificateDomainArray{
			&acme.CertificateDomainArgs{Domain: pulumi.String(config.C.Host(pcfg.Host)), Plugin: plugin.Plugin},
		},
	}, pulumi.Provider(provider))
	return err
}

// NewProxmoxSSO adds an "authentik" login realm. Users are created on first
// login from their GitHub username, and Authentik's groups sync in as
// <group>-authentik, so grafana-admins members get Administrator. root@pam
// remains the fallback login.
func NewProxmoxSSO(ctx *pulumi.Context, provider *proxmoxve.Provider) error {
	clientSecret := cfg.K.String("PROXMOX_OAUTH_CLIENT_SECRET")
	if clientSecret == "" {
		return fmt.Errorf("PROXMOX_OAUTH_CLIENT_SECRET not found in .env")
	}

	_, err := realm.NewOpenid(ctx, "pve-realm-authentik", &realm.OpenidArgs{
		Realm:            pulumi.String("authentik"),
		Comment:          pulumi.String("Authentik (GitHub)"),
		IssuerUrl:        pulumi.String("https://" + config.C.Host("auth") + "/application/o/proxmox/"),
		ClientId:         pulumi.String("proxmox-homelab"),
		ClientKey:        pulumi.ToSecret(pulumi.String(clientSecret)).(pulumi.StringOutput),
		Scopes:           pulumi.String("openid email profile groups"),
		UsernameClaim:    pulumi.String("preferred_username"),
		Autocreate:       pulumi.Bool(true),
		GroupsClaim:      pulumi.String("groups"),
		GroupsAutocreate: pulumi.Bool(true),
		GroupsOverwrite:  pulumi.Bool(true),
		Default:          pulumi.Bool(true),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	_, err = proxmoxve.NewGroupLegacy(ctx, "pve-group-grafana-admins", &proxmoxve.GroupLegacyArgs{
		GroupId: pulumi.String("grafana-admins-authentik"),
		Comment: pulumi.String("Synced from Authentik group grafana-admins"),
		Acls: proxmoxve.GroupLegacyAclArray{ //nolint:staticcheck // moving to the Acl resource would replace live state
			&proxmoxve.GroupLegacyAclArgs{Path: pulumi.String("/"), RoleId: pulumi.String("Administrator"), Propagate: pulumi.Bool(true)},
		},
	}, pulumi.Provider(provider))
	return err
}

// DownloadImage downloads a disk image to the Proxmox "local" datastore.
// A replacement (new URL or schematic) deletes the old file first, since Proxmox
// refuses to overwrite an existing one; OverwriteUnmanaged adopts a same-named
// file that is not in Pulumi state.
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
// It uses OVMF + q35 with an EFI disk, and passes each of the node's PcieIDs
// through as a PCIe device. VMs are protected: control planes are etcd members.
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
				// Compute-only GPU; primary-VGA mode adds BAR constraints that conflict
				// with large memory allocations.
				Xvga: pulumi.Bool(false),
			})
		}
		args.Hostpcis = hostpcis
	}

	// Unprotect explicitly for an intentional rebuild.
	return proxmoxve.NewVmLegacy(ctx, node.Name, args,
		pulumi.Provider(provider),
		pulumi.Protect(true),
		// The image only seeds a new VM's disk; a newer image must not replace
		// running nodes, which upgrade in place via `just talos-upgrade`.
		// The provider reads an unset speed block back as null, which would
		// otherwise plan a no-op disk update on every run.
		pulumi.IgnoreChanges([]string{"disks[0].fileId", "disks[0].speed"}),
	)
}
