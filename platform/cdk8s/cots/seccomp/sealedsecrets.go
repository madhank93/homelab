package seccomp

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewSealedSecretsChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	cdk8s.NewHelm(chart, jsii.String("sealed-secrets"), &cdk8s.HelmProps{
		Chart:       jsii.String("sealed-secrets"),
		Namespace:   jsii.String("kube-system"),
		Repo:        jsii.String("https://bitnami-labs.github.io/sealed-secrets"),
		ReleaseName: jsii.String("sealed-secrets-controller"),
		Version:     jsii.String("1.16.0"), // Latest stable chart
		Values: &map[string]any{
			"fullnameOverride": "sealed-secrets-controller",
			"image": map[string]any{
				"registry":   "ghcr.io",
				"repository": "bitnami-labs/sealed-secrets-controller",
				"tag":        "0.34.0", // Tag on GHCR is without 'v'
			},
		},
	})
	return chart
}
