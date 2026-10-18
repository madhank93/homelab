package cloud

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/core/internal/cfg"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	hcloud "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DeployHetznerVPS provisions the Bifrost VPS on Hetzner Cloud and bootstraps it.
//
// It creates a firewall (SSH 22, Traefik 80/443, NetBird STUN 3478/udp), a server
// with cloud-init, then copies the ./cloud/bifrost/ config directory to the server
// and runs bootstrap.sh. The bootstrap is re-triggered whenever configs or secrets
// change via a SHA-256 hash stored as a Pulumi trigger.
//
// Secrets are written to one env file per consumer under ./cloud/bifrost/env/
// (gitignored) before the copy step. Run via `just core hetzner up`.
func DeployHetznerVPS(ctx *pulumi.Context) error {
	hcfg := config.C.Hetzner

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
			pulumi.String(hcfg.SSHKey),
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

	// Env files are gitignored and overwritten on every pulumi up — never edit by hand.
	if err := os.MkdirAll(bifrostEnvDir, 0700); err != nil {
		return err
	}
	for _, f := range bifrostEnvFiles {
		path := filepath.Join(bifrostEnvDir, f.name+".env")
		if err := writeBifrostEnv(path, f.vars); err != nil {
			return fmt.Errorf("generate bifrost %s: %w", path, err)
		}
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

type secretVar struct {
	envKey  string // name in the SOPS environment
	fileKey string // name written to the env file
	require bool   // hard-fail if empty
}

const bifrostEnvDir = "./cloud/bifrost/env"

var (
	// Authentik Django secret key — generate: openssl rand -base64 48
	authentikSecretKey = secretVar{"AUTHENTIK_SECRET_KEY", "AUTHENTIK_SECRET_KEY", true}
	// Authentik Postgres password — generate: openssl rand -hex 16
	authentikDBPassword = secretVar{"AUTHENTIK_POSTGRESQL_PASSWORD", "AUTHENTIK_POSTGRESQL__PASSWORD", true}
)

// bifrostEnvFiles lists env/<name>.env and the secrets it holds. Each file has
// exactly one consumer (a compose service, or bootstrap.sh itself), so a
// container never sees another service's secrets.
var bifrostEnvFiles = []struct {
	name string
	vars []secretVar
}{
	// Cloudflare ACME DNS challenge provider.
	{"traefik", []secretVar{{"CLOUDFLARE_API_TOKEN", "CF_DNS_API_TOKEN", true}}},

	{"authentik-postgres", []secretVar{{"AUTHENTIK_POSTGRESQL_PASSWORD", "POSTGRES_PASSWORD", true}}},

	// The bootstrap token becomes the akadmin API token on first boot; the same
	// SOPS value is reused across fresh deployments.
	{"authentik-server", []secretVar{
		authentikSecretKey,
		authentikDBPassword,
		{"AUTHENTIK_TOKEN", "AUTHENTIK_BOOTSTRAP_TOKEN", true},
	}},

	{"authentik-worker", []secretVar{authentikSecretKey, authentikDBPassword}},

	// Personal Access Token created in the NetBird UI after first login;
	// bootstrap.sh skips netbird-proxy until it is set.
	{"netbird-proxy", []secretVar{{"NB_PROXY_TOKEN", "NB_PROXY_TOKEN", false}}},

	// Reusable setup key (NetBird UI → Setup Keys) for the bifrost routing peer.
	{"netbird-agent", []secretVar{{"NB_BIFROST_SETUP_KEY", "NB_SETUP_KEY", true}}},

	// Substituted into netbird/config.yaml by bootstrap.sh (NetBird does not
	// expand env vars there). Keys: openssl rand -base64 32; the owner password
	// is bcrypt-hashed for the embedded Dex admin account.
	{"bootstrap", []secretVar{
		{"NB_DATA_STORE_KEY", "NB_DATA_STORE_KEY", true},
		{"NB_RELAY_SECRET", "NB_RELAY_SECRET", true},
		{"NB_OWNER_PASSWORD", "NB_OWNER_PASSWORD", true},
	}},
}

// writeBifrostEnv writes vars to path as KEY=value lines. Values come from the
// SOPS-injected environment (sops exec-env); the file is gitignored and
// overwritten on every pulumi up.
func writeBifrostEnv(path string, vars []secretVar) error {
	log.Printf("[hetzner] generating %s", filepath.Base(path))

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

	if err := os.WriteFile(path, []byte(sb.String()), 0600); err != nil {
		return err
	}
	log.Printf("[hetzner]   wrote %s", path)
	return nil
}

// computeBifrostHash returns a SHA-256 of every file under ./cloud/bifrost, the
// generated env files included. Used as a Pulumi Trigger so bootstrap.sh re-runs
// when either configs or secrets change. Values are hashed — never stored.
func computeBifrostHash() (string, error) {
	h := sha256.New()

	err := filepath.Walk("./cloud/bifrost", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
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
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// generateTraefikPublicServices writes ./cloud/bifrost/traefik/dynamic/public-services.yml.
// The file is gitignored and regenerated on every pulumi up. Traefik's file provider
// hot-reloads the new routes without a container restart.
func generateTraefikPublicServices(services []PublicService) error {
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
		sb.WriteString(fmt.Sprintf("    %s:\n", svc.Name))
		sb.WriteString(fmt.Sprintf("      rule: \"Host(`%s`)\"\n", config.C.Host(svc.Name)))
		if !svc.SkipAuth {
			sb.WriteString("      middlewares: [authentik-forwardauth]\n")
		}
		sb.WriteString("      service: k8s-gateway\n")
		sb.WriteString("      tls:\n")
		sb.WriteString("        certResolver: cloudflare-dns\n")
	}

	return os.WriteFile(fmt.Sprintf("%s/public-services.yml", dir), []byte(sb.String()), 0644)
}
