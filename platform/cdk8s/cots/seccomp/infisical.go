package seccomp

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	infisical "github.com/madhank93/homelab/cdk8s/imports/infisicalstandalone"
)

func NewInfisicalChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{
		"frontend": map[string]any{
			"enabled": true,
			"service": map[string]any{
				"type": "ClusterIP",
				"port": 3000,
			},
		},
		"backend": map[string]any{
			"enabled": true,
			"service": map[string]any{
				"type": "ClusterIP",
				"port": 4000,
			},
		},
		"postgresql": map[string]any{
			"enabled": true,
			"auth": map[string]any{
				"database": "infisical",
				"username": "infisical",
				"password": "infisical123",
			},
			"primary": map[string]any{
				"persistence": map[string]any{
					"size": "20Gi",
				},
			},
		},
		"redis": map[string]any{
			"enabled": true,
			"auth": map[string]any{
				"enabled": false,
			},
		},
		"ingress": map[string]any{
			"enabled":  true,
			"hostname": "infisical.local",
		},
	}

	infisical.NewInfisicalstandalone(chart, jsii.String("infisical-release"), &infisical.InfisicalstandaloneProps{
		ReleaseName: jsii.String("infisical"),
		Values:      &values,
	})

	return chart
}
