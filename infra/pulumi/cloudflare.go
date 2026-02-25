package main

import (
	"os"

	cf "github.com/pulumi/pulumi-cloudflare/sdk/v5/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ManageCloudflare manages Cloudflare DNS records for the homelab.
func ManageCloudflare(ctx *pulumi.Context) error {
	provider, err := cf.NewProvider(ctx, "cloudflare", &cf.ProviderArgs{
		ApiToken: pulumi.String(os.Getenv("CLOUDFLARE_API_TOKEN")),
	})
	if err != nil {
		return err
	}

	// Zone ID for madhan.app
	zoneID := "ab8d029e91fb2bb38cb8bc91b9f0b218"

	// Wildcard A record — routes all *.madhan.app to the homelab gateway LAN IP
	_, err = cf.NewRecord(ctx, "wildcard-madhan-app", &cf.RecordArgs{
		ZoneId:  pulumi.String(zoneID),
		Name:    pulumi.String("*"),
		Type:    pulumi.String("A"),
		Content: pulumi.String("192.168.1.220"),
		Ttl:     pulumi.Int(1), // Auto TTL
		Proxied: pulumi.Bool(false),
		Comment: pulumi.StringPtr("Homelab wildcard - LAN gateway"),
	}, pulumi.Provider(provider))

	return err
}
