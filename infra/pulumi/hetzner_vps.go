package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"

	hcloud "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type HetznerConfig struct {
	ServerName string `koanf:"server_name"`
	Image      string `koanf:"image"`
	ServerType string `koanf:"server_type"`
	Location   string `koanf:"location"`
	SshKey     string `koanf:"ssh_key"`
	VpsIP      string `koanf:"vps_ip"`
}

func DeployHetznerVPS(ctx *pulumi.Context) error {
	var cfg HetznerConfig
	if err := LoadConfig("hetzner", &cfg); err != nil {
		return err
	}

	token := k.String("HCLOUD_TOKEN")
	if token == "" {
		return fmt.Errorf("HCLOUD_TOKEN not found; make sure it's in your environment")
	}

	provider, err := hcloud.NewProvider(ctx, "hcloud", &hcloud.ProviderArgs{
		Token: pulumi.StringPtr(token),
	})
	if err != nil {
		return err
	}

	fw, err := hcloud.NewFirewall(ctx, "bifrost-fw", &hcloud.FirewallArgs{
		Name: pulumi.String("bifrost-fw"),
		Rules: hcloud.FirewallRuleArray{
			&hcloud.FirewallRuleArgs{
				Direction: pulumi.String("in"),
				Protocol:  pulumi.String("tcp"),
				Port:      pulumi.String("22"),
				SourceIps: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
					pulumi.String("::/0"),
				},
			},
			&hcloud.FirewallRuleArgs{
				Direction: pulumi.String("in"),
				Protocol:  pulumi.String("tcp"),
				Port:      pulumi.String("443"),
				SourceIps: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
					pulumi.String("::/0"),
				},
			},
			&hcloud.FirewallRuleArgs{
				Direction: pulumi.String("in"),
				Protocol:  pulumi.String("tcp"),
				Port:      pulumi.String("80"),
				SourceIps: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
					pulumi.String("::/0"),
				},
			},
			&hcloud.FirewallRuleArgs{
				Direction: pulumi.String("in"),
				Protocol:  pulumi.String("udp"),
				Port:      pulumi.String("3478"),
				SourceIps: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
					pulumi.String("::/0"),
				},
			},
			&hcloud.FirewallRuleArgs{
				Direction: pulumi.String("in"),
				Protocol:  pulumi.String("tcp"),
				Port:      pulumi.String("3478"),
				SourceIps: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
					pulumi.String("::/0"),
				},
			},
			&hcloud.FirewallRuleArgs{
				Direction: pulumi.String("in"),
				Protocol:  pulumi.String("udp"),
				Port:      pulumi.String("5349"),
				SourceIps: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
					pulumi.String("::/0"),
				},
			},
			&hcloud.FirewallRuleArgs{
				Direction: pulumi.String("in"),
				Protocol:  pulumi.String("tcp"),
				Port:      pulumi.String("5349"),
				SourceIps: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
					pulumi.String("::/0"),
				},
			},
			&hcloud.FirewallRuleArgs{
				Direction: pulumi.String("in"),
				Protocol:  pulumi.String("udp"),
				Port:      pulumi.String("50000-50500"), // TURN Ephemeral Range
				SourceIps: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
					pulumi.String("::/0"),
				},
			},
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	fwID := fw.ID().ApplyT(func(id string) (int, error) {
		i, err := strconv.Atoi(id)
		return i, err
	}).(pulumi.IntOutput)

	userData, err := os.ReadFile("./cloud-init/cloud-init-hetzner.yml")
	if err != nil {
		return err
	}

	server, err := hcloud.NewServer(ctx, "hetzner-vps", &hcloud.ServerArgs{
		Name:       pulumi.String(cfg.ServerName),
		Image:      pulumi.String(cfg.Image),
		ServerType: pulumi.String(cfg.ServerType),
		Location:   pulumi.String(cfg.Location),
		SshKeys: pulumi.StringArray{
			pulumi.String(cfg.SshKey),
		},
		FirewallIds: pulumi.IntArray{
			fwID,
		},
		PublicNets: hcloud.ServerPublicNetArray{
			&hcloud.ServerPublicNetArgs{
				Ipv4Enabled: pulumi.Bool(true),
				Ipv6Enabled: pulumi.Bool(true),
			},
		},
		UserData: pulumi.String(string(userData)),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	serverIP := server.Ipv4Address

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	keyPath := filepath.Join(home, ".ssh", "id_ed25519")
	privateKey, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}

	// Generate public-services.yml so Traefik file-watcher hot-reloads internet-exposed routes.
	// publicServices is defined in cloudflare.go (same package). Edit that list to toggle exposure.
	if err := generateTraefikPublicServices(publicServices); err != nil {
		return fmt.Errorf("failed to generate traefik public-services.yml: %w", err)
	}

	// Inject CF_DNS_API_TOKEN (Traefik Cloudflare ACME) from the same CLOUDFLARE_API_TOKEN
	// that cert-manager already uses. Same token, different env var name required by Traefik.
	// Value is available because pulumi runs under: sops exec-env bootstrap.sops.yaml
	if err := generateBifrostSecretsEnv(); err != nil {
		return fmt.Errorf("failed to generate bifrost .secrets.env: %w", err)
	}

	conn := &remote.ConnectionArgs{
		Host:       serverIP,
		User:       pulumi.String("root"),
		PrivateKey: pulumi.String(string(privateKey)),
	}

	_, err = remote.NewCopyToRemote(ctx, "configs-dir",
		&remote.CopyToRemoteArgs{
			Connection: conn,
			Source:     pulumi.NewFileArchive("./bifrost"),
			RemotePath: pulumi.String("/etc"),
		}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	return nil
}

// generateBifrostSecretsEnv writes ./bifrost/.secrets.env with secrets that are available
// in the pulumi environment (injected by sops exec-env) but need different names on the VPS.
// The file is gitignored by .env* and copied to the VPS alongside .env.
func generateBifrostSecretsEnv() error {
	// CLOUDFLARE_API_TOKEN (from bootstrap SOPS, used by cert-manager) →
	// CF_DNS_API_TOKEN (required by Traefik's Cloudflare ACME DNS challenge provider)
	cfToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	if cfToken == "" {
		return fmt.Errorf("CLOUDFLARE_API_TOKEN not in environment; run via: just pulumi hetzner up")
	}

	content := fmt.Sprintf("# Auto-generated by hetzner_vps.go — do not edit or commit\nCF_DNS_API_TOKEN=%s\n", cfToken)
	return os.WriteFile("./bifrost/.secrets.env", []byte(content), 0600)
}

// generateTraefikPublicServices writes ./bifrost/traefik/dynamic/public-services.yml.
// The file is gitignored and regenerated on every pulumi up. Traefik's file provider
// hot-reloads the new routes without a container restart.
func generateTraefikPublicServices(services []string) error {
	dir := "./bifrost/traefik/dynamic"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create traefik dynamic dir: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("# Auto-generated by hetzner_vps.go — do not edit manually.\n")
	sb.WriteString("# Edit publicServices in cloudflare.go and re-run: just pulumi hetzner up\n\n")
	sb.WriteString("http:\n")
	sb.WriteString("  routers:\n")
	for _, svc := range services {
		sb.WriteString(fmt.Sprintf("    %s:\n", svc))
		sb.WriteString(fmt.Sprintf("      rule: \"Host(`%s.madhan.app`)\"\n", svc))
		sb.WriteString("      middlewares: [authentik-forwardauth]\n")
		sb.WriteString("      service: k8s-gateway\n")
		sb.WriteString("      tls:\n")
		sb.WriteString("        certResolver: cloudflare-dns\n")
	}

	outPath := fmt.Sprintf("%s/public-services.yml", dir)
	return os.WriteFile(outPath, []byte(sb.String()), 0644)
}
