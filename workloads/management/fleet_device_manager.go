package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/fleet"
)

func NewFleetChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	fleet.NewFleet(chart, jsii.String("fleet-release"), &fleet.FleetProps{
		ReleaseName: jsii.String("fleet"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]interface{}{
			"fleet":  map[string]interface{}{"enabled": true},
			"gitops": map[string]interface{}{"enabled": true},
			"controller": map[string]interface{}{
				"replicas": 1,
				"resources": map[string]interface{}{
					"limits":   map[string]interface{}{"memory": "512Mi", "cpu": "500m"},
					"requests": map[string]interface{}{"memory": "256Mi", "cpu": "100m"},
				},
			},
			"agent":                 map[string]interface{}{"enabled": true},
			"debug":                 false,
			"systemDefaultRegistry": "",
		},
	})

	return chart
}
