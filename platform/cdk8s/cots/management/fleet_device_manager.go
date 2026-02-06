package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/fleet"
)

func NewFleetChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{})

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

	fleet.NewFleet(chart, jsii.String("fleet-release"), &fleet.FleetProps{
		ReleaseName: jsii.String("fleet"),
		Values:      &values,
	})

	return chart
}
