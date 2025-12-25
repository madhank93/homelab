package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"

	hcloud "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployHetznerVPS(ctx *pulumi.Context) error {
	token := os.Getenv("HCLOUD_TOKEN")
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
	})
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
		Name:       pulumi.String("bifrost-public-vps1"),
		Image:      pulumi.String("ubuntu-24.04"),
		ServerType: pulumi.String("cpx21"),
		Location:   pulumi.String("ash"),
		SshKeys: pulumi.StringArray{
			pulumi.String("mac-ssh"),
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

	conn := &remote.ConnectionArgs{
		Host: serverIP,
		User: pulumi.String("root"),
	}

	copyDir, err := remote.NewCopyToRemote(ctx, "configs-dir",
		&remote.CopyToRemoteArgs{
			Connection: conn,
			Source:     pulumi.NewFileArchive("./bifrost"),
			RemotePath: pulumi.String("/etc"),
		})
	if err != nil {
		return err
	}

	_, err = remote.NewCommand(ctx, "run-setup", &remote.CommandArgs{
		Connection: conn,
		Create:     pulumi.String("bash /etc/bifrost/bootstrap.sh"),
	}, pulumi.DependsOn([]pulumi.Resource{copyDir}))
	if err != nil {
		return err
	}

	return nil
}
