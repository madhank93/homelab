package seccomp

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewCertManagerChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{
		"installCRDs": true,
		"global": map[string]any{
			"leaderElection": map[string]any{
				"namespace": namespace,
			},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("cert-manager"), &cdk8s.HelmProps{
		Chart:       jsii.String("cert-manager"),
		Repo:        jsii.String("https://charts.jetstack.io"),
		Version:     jsii.String("v1.16.2"),
		ReleaseName: jsii.String("cert-manager"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
