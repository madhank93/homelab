package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewFleetChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{
		"fleet": map[string]any{
			"enabled": true,
		},
		"gitops": map[string]any{
			"enabled": true,
		},
		"controller": map[string]any{
			"replicas": 1,
			"resources": map[string]any{
				"limits": map[string]any{
					"memory": "512Mi",
					"cpu":    "500m",
				},
				"requests": map[string]any{
					"memory": "256Mi",
					"cpu":    "100m",
				},
			},
		},
		"agent": map[string]any{
			"enabled": true,
		},
		"debug":                 false,
		"systemDefaultRegistry": "",
	}

	cdk8s.NewHelm(chart, jsii.String("fleet-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("fleet"),
		Repo:        jsii.String("https://rancher.github.io/fleet-helm-charts"),
		Version:     jsii.String("0.14.2"),
		ReleaseName: jsii.String("fleet"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
