package main

import (
	"fmt"
	"os"

	"github.com/madhank93/homelab/core/cloud"
	"github.com/madhank93/homelab/core/internal/cfg"
	"github.com/madhank93/homelab/core/platform"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	if err := cfg.Init(); err != nil {
		fmt.Printf("Critical: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	for _, key := range cfg.K.Keys() {
		if err := os.Setenv(key, cfg.K.String(key)); err != nil {
			fmt.Printf("Critical: Failed to export %s: %v\n", key, err)
			os.Exit(1)
		}
	}

	pulumi.Run(func(ctx *pulumi.Context) error {
		switch ctx.Stack() {
		case "talos":
			return platform.DeployTalosCluster(ctx)
		case "platform":
			return platform.DeployPlatform(ctx)
		case "hetzner":
			return cloud.DeployHetznerVPS(ctx)
		case "authentik":
			return cloud.DeployAuthentik(ctx)
		case "cloudflare":
			return cloud.ManageCloudflare(ctx)
		default:
			return fmt.Errorf("unknown stack: %q — valid stacks: talos, platform, hetzner, authentik, cloudflare", ctx.Stack())
		}
	})
}
