package cloud

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/madhank93/homelab/core/internal/cfg"
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
	var hcfg HetznerConfig
	if err := cfg.Load("hetzner", &hcfg); err != nil {
		return err
	}

	token := cfg.K.String("HCLOUD_TOKEN")
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
				Port:      pulumi.String("50000-50500"), // TURN ephemeral range
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

	userData, err := os.ReadFile("./cloud/cloud-init/cloud-init-hetzner.yml")
	if err != nil {
		return err
	}

	server, err := hcloud.NewServer(ctx, "hetzner-vps", &hcloud.ServerArgs{
		Name:       pulumi.String(hcfg.ServerName),
		Image:      pulumi.String(hcfg.Image),
		ServerType: pulumi.String(hcfg.ServerType),
		Location:   pulumi.String(hcfg.Location),
		SshKeys: pulumi.StringArray{
			pulumi.String(hcfg.SshKey),
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

	// Generate public-services.yml so Traefik hot-reloads internet-exposed routes.
	// publicServices is defined in cloudflare.go. Edit that slice to toggle exposure.
	if err := generateTraefikPublicServices(publicServices); err != nil {
		return fmt.Errorf("generate traefik public-services.yml: %w", err)
	}

	// Write .secrets.env and .env with all VPS secrets from the SOPS-injected environment.
	// Both files are gitignored and overwritten on every pulumi up — never edit by hand.
	if err := generateBifrostSecretsEnv(); err != nil {
		return fmt.Errorf("generate bifrost .secrets.env: %w", err)
	}
	if err := generateBifrostDotEnv(); err != nil {
		return fmt.Errorf("generate bifrost .env: %w", err)
	}

	// Trigger hash — bootstrap.sh re-runs when configs or secrets change.
	triggerHash, err := computeBifrostHash()
	if err != nil {
		return fmt.Errorf("hash bifrost dir: %w", err)
	}

	conn := &remote.ConnectionArgs{
		Host:       serverIP,
		User:       pulumi.String("root"),
		PrivateKey: pulumi.String(string(privateKey)),
	}

	copyResult, err := remote.NewCopyToRemote(ctx, "configs-dir",
		&remote.CopyToRemoteArgs{
			Connection: conn,
			Source:     pulumi.NewFileArchive("./cloud/bifrost"),
			RemotePath: pulumi.String("/etc"),
		}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	// Run bootstrap.sh on the VPS after every config/secret change.
	// The script starts services in order, waits for health, and logs each step.
	_, err = remote.NewCommand(ctx, "bifrost-bootstrap",
		&remote.CommandArgs{
			Connection: conn,
			Create:     pulumi.String("bash /etc/bifrost/bootstrap.sh"),
			Update:     pulumi.String("bash /etc/bifrost/bootstrap.sh"),
			Triggers:   pulumi.Array{pulumi.String(triggerHash)},
		},
		pulumi.DependsOn([]pulumi.Resource{copyResult}),
	)
	return err
}

// generateBifrostSecretsEnv writes ./cloud/bifrost/.secrets.env with all secrets the
// bifrost stack needs. Values come from the SOPS-injected environment (sops exec-env).
// The file is gitignored and overwritten on every pulumi up.
func generateBifrostSecretsEnv() error {
	type secretVar struct {
		envKey  string // name in the SOPS environment
		fileKey string // name written to .secrets.env
		require bool   // hard-fail if empty
	}

	vars := []secretVar{
		// Traefik Cloudflare ACME DNS challenge provider
		{"CLOUDFLARE_API_TOKEN", "CF_DNS_API_TOKEN", true},

		// NetBird datastore encryption key — generate: openssl rand -base64 32
		{"NB_DATA_STORE_KEY", "NB_DATA_STORE_KEY", true},

		// NetBird relay auth secret — generate: openssl rand -base64 32
		{"NB_RELAY_SECRET", "NB_RELAY_SECRET", true},

		// Authentik bootstrap token — Authentik creates this as the akadmin API token
		// on first boot (AUTHENTIK_BOOTSTRAP_TOKEN env var). bootstrap.sh then reads it
		// to call the Authentik API and writes it as NB_IDP_MGMT_TOKEN for NetBird.
		// Same SOPS value reused across fresh deployments.
		{"AUTHENTIK_TOKEN", "AUTHENTIK_BOOTSTRAP_TOKEN", true},

		// NetBird Personal Access Token for netbird-proxy.
		// Created in the NetBird UI after first login. Optional on initial deploy —
		// bootstrap.sh skips netbird-proxy until this is present.
		{"NB_PROXY_TOKEN", "NB_PROXY_TOKEN", false},

		// NetBird setup key for the bifrost WireGuard routing agent.
		// Created in the NetBird UI under Setup Keys. Optional on initial deploy.
		{"NB_BIFROST_SETUP_KEY", "NB_BIFROST_SETUP_KEY", false},
	}

	log.Printf("[hetzner] generating .secrets.env")

	var sb strings.Builder
	sb.WriteString("# Auto-generated by hetzner.go — do not edit or commit\n")
	sb.WriteString("# Regenerated on every: just core hetzner up\n\n")

	for _, v := range vars {
		val := os.Getenv(v.envKey)
		if val == "" {
			if v.require {
				return fmt.Errorf(
					"required secret %s not set; add it to secrets/bootstrap.sops.yaml",
					v.envKey,
				)
			}
			log.Printf("[hetzner]   SKIP  %-22s  (not in SOPS — optional)", v.fileKey)
			continue
		}
		sb.WriteString(fmt.Sprintf("%s=%s\n", v.fileKey, val))
		log.Printf("[hetzner]   SET   %-22s  (%d chars)", v.fileKey, len(val))
	}

	path := "./cloud/bifrost/.secrets.env"
	if err := os.WriteFile(path, []byte(sb.String()), 0600); err != nil {
		return err
	}
	log.Printf("[hetzner]   wrote %s", path)
	return nil
}

// generateBifrostDotEnv writes ./cloud/bifrost/.env with Authentik secrets.
// Docker Compose reads this file for ${VAR} interpolation within docker-compose.yml
// (e.g. POSTGRES_PASSWORD=${AUTHENTIK_POSTGRESQL_PASSWORD}) — it must be named .env
// in the project directory and cannot be merged into .secrets.env for that purpose.
func generateBifrostDotEnv() error {
	type secretVar struct {
		envKey  string
		fileKey string
		require bool
	}

	vars := []secretVar{
		// Authentik Django secret key — generate: openssl rand -base64 48
		{"AUTHENTIK_SECRET_KEY", "AUTHENTIK_SECRET_KEY", true},

		// Authentik Postgres password — generate: openssl rand -hex 16
		{"AUTHENTIK_POSTGRESQL_PASSWORD", "AUTHENTIK_POSTGRESQL_PASSWORD", true},
	}

	log.Printf("[hetzner] generating .env")

	var sb strings.Builder
	sb.WriteString("# Auto-generated by hetzner.go — do not edit or commit\n")
	sb.WriteString("# Regenerated on every: just core hetzner up\n\n")

	for _, v := range vars {
		val := os.Getenv(v.envKey)
		if val == "" {
			if v.require {
				return fmt.Errorf(
					"required secret %s not set; add it to secrets/bootstrap.sops.yaml",
					v.envKey,
				)
			}
			log.Printf("[hetzner]   SKIP  %-30s  (not in SOPS — optional)", v.fileKey)
			continue
		}
		sb.WriteString(fmt.Sprintf("%s=%s\n", v.fileKey, val))
		log.Printf("[hetzner]   SET   %-30s  (%d chars)", v.fileKey, len(val))
	}

	path := "./cloud/bifrost/.env"
	if err := os.WriteFile(path, []byte(sb.String()), 0600); err != nil {
		return err
	}
	log.Printf("[hetzner]   wrote %s", path)
	return nil
}

// computeBifrostHash returns a SHA-256 of all bifrost config files (excluding secret
// files) plus secret key=value pairs. Used as a Pulumi Trigger so bootstrap.sh
// re-runs when either configs or secrets change. Values are hashed — never stored.
func computeBifrostHash() (string, error) {
	h := sha256.New()

	skipFiles := map[string]bool{
		".secrets.env": true,
		".env":         true,
	}

	err := filepath.Walk("./cloud/bifrost", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if skipFiles[filepath.Base(path)] {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		h.Write([]byte(path))
		h.Write(data)
		return nil
	})
	if err != nil {
		return "", err
	}

	// Mix in secret values so the hash changes on secret rotation.
	for _, k := range []string{
		"CLOUDFLARE_API_TOKEN", "NB_DATA_STORE_KEY", "NB_RELAY_SECRET",
		"AUTHENTIK_TOKEN", "NB_PROXY_TOKEN", "NB_BIFROST_SETUP_KEY",
		"AUTHENTIK_SECRET_KEY", "AUTHENTIK_POSTGRESQL_PASSWORD",
	} {
		h.Write([]byte(k + "=" + os.Getenv(k) + "\n"))
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// generateTraefikPublicServices writes ./cloud/bifrost/traefik/dynamic/public-services.yml.
// The file is gitignored and regenerated on every pulumi up. Traefik's file provider
// hot-reloads the new routes without a container restart.
func generateTraefikPublicServices(services []string) error {
	dir := "./cloud/bifrost/traefik/dynamic"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create traefik dynamic dir: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("# Auto-generated by hetzner.go — do not edit manually.\n")
	sb.WriteString("# Edit publicServices in cloudflare.go and re-run: just core hetzner up\n\n")
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

	return os.WriteFile(fmt.Sprintf("%s/public-services.yml", dir), []byte(sb.String()), 0644)
}
