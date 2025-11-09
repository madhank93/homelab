package main

import (
	"fmt"
	"os"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var k = koanf.New(".")

func main() {
	// Load .env globally for all modules
	if err := k.Load(file.Provider(".env"), dotenv.Parser()); err != nil {
		fmt.Println("Warning: could not load .env:", err)
	}

	for _, key := range k.Keys() {
		val := k.String(key)
		os.Setenv(key, val)
	}

	pulumi.Run(func(ctx *pulumi.Context) error {
		// Deploy Hetzner VPS
		if err := DeployHetznerVPS(ctx); err != nil {
			return fmt.Errorf("hetzner VPS deployment failed: %v", err)
		}

		// Deploy Proxmox Homelab
		// if err := DeployHomelab(ctx); err != nil {
		// 	return fmt.Errorf("homelab deployment failed: %v", err)
		// }

		return nil
	})
}
