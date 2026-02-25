package main

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type service struct {
	configKey string
	deploy    func(ctx *pulumi.Context) error
}

func main() {
	// Load .env globally for all modules
	if err := InitConfig(); err != nil {
		fmt.Printf("Critical: Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	for _, key := range k.Keys() {
		val := k.String(key)
		os.Setenv(key, val)
	}

	services := map[string]service{
		"hetzner": {
			configKey: "services.hetzner.enabled",
			deploy:    DeployHetznerVPS,
		},
		"talos": {
			configKey: "services.talos.enabled",
			deploy:    DeployTalosCluster,
		},
		"authentik": {
			configKey: "services.authentik.enabled",
			deploy:    DeployAuthentik,
		},
		"cloudflare": {
			configKey: "services.cloudflare.enabled",
			deploy:    ManageCloudflare,
		},
	}

	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "") // project namespace

		for name, s := range services {
			enabled := false

			if v, err := cfg.TryBool(s.configKey); err == nil {
				enabled = v
			}

			if !enabled {
				fmt.Println("Skipping service:", name)
				continue
			}

			if err := s.deploy(ctx); err != nil {
				return fmt.Errorf("%s deployment failed: %v", name, err)
			}
		}

		return nil
	})
}
