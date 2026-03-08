package secrets

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewCsiDriverChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String("kube-system"),
	})

	// Secrets Store CSI Driver — manages the CSI interface that allows pods to
	// mount secrets from external providers (OpenBao) as files. Must be deployed
	// in kube-system with tolerations covering all nodes so every node can serve
	// CSI mount requests.
	// syncSecret.enabled=true: creates k8s Secrets from secretObjects in SecretProviderClass,
	// required for apps that use existingSecret (Harbor, N8n, Rancher, NetBird).
	cdk8s.NewHelm(chart, jsii.String("csi-driver-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("secrets-store-csi-driver"),
		Repo:        jsii.String("https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts"),
		Version:     jsii.String("1.5.6"),
		ReleaseName: jsii.String("secrets-store-csi-driver"),
		Namespace:   jsii.String("kube-system"),
		Values: &map[string]any{
			// Required for Pattern B (Harbor, N8n, Rancher, NetBird) — creates k8s Secret
			// from secretObjects defined in SecretProviderClass when a pod mounts the CSI volume.
			"syncSecret": map[string]any{
				"enabled": true,
			},
			"enableSecretRotation": true,
			"rotationPollInterval": "2m",
			"linux": map[string]any{
				// Tolerate all nodes (control-plane taint included) so every node
				// in the cluster can serve CSI mount requests for pods it hosts.
				"tolerations": []map[string]any{
					{"operator": "Exists"},
				},
			},
		},
	})

	return chart
}
