package secrets

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
)

// NewCsiDriverChart deploys the Secrets Store CSI Driver into kube-system.
// The DaemonSet tolerates every node so any node can serve CSI mounts for the
// pods it hosts.
func NewCsiDriverChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String("kube-system"),
	})

	csi := config.HelmChart("secrets-store-csi-driver")
	cdk8s.NewHelm(chart, jsii.String("csi-driver-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("secrets-store-csi-driver"),
		Repo:        jsii.String(csi.Repo),
		Version:     jsii.String(csi.Version),
		ReleaseName: jsii.String("secrets-store-csi-driver"),
		Namespace:   jsii.String("kube-system"),
		Values: &map[string]any{
			// Mirrors SecretProviderClass secretObjects into k8s Secrets, for
			// apps that read an existingSecret.
			"syncSecret": map[string]any{
				"enabled": true,
			},
			"enableSecretRotation": true,
			"rotationPollInterval": "2m",
			"linux": map[string]any{
				"tolerations": []map[string]any{
					{"operator": "Exists"},
				},
			},
		},
	})

	return chart
}
