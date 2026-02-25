package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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
