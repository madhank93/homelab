package main

import (
	"fmt"
	"os"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

var k = koanf.New(".")

type service struct {
	configKey string
	deploy    func(ctx *pulumi.Context) error
}

func main() {
	// Load .env globally for all modules
	if err := k.Load(file.Provider(".env"), dotenv.Parser()); err != nil {
		fmt.Println("Warning: could not load .env:", err)
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
		"proxmox": {
			configKey: "services.proxmox.enabled",
			deploy:    DeployProxmox,
		},
	}

	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "") // project namespace

		for name, s := range services {
			enabled := true

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
