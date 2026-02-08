package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewHeadlampChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	cdk8s.NewHelm(chart, jsii.String("headlamp-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("headlamp"),
		Repo:        jsii.String("https://kubernetes-sigs.github.io/headlamp/"),
		ReleaseName: jsii.String("headlamp"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
			"ingress": map[string]any{
				"enabled": true,
				"hosts": []map[string]any{
					{
						"host": "headlamp.madhan.app",
						"paths": []map[string]any{
							{"path": "/", "type": "ImplementationSpecific"},
						},
					},
				},
				"tls": []map[string]any{
					{
						"secretName": "headlamp-tls",
						"hosts":      []any{"headlamp.madhan.app"},
					},
				},
			},
		},
	})
	return chart
}
