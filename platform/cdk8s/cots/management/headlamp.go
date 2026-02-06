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
		Values: &map[string]interface{}{
			"ingress": map[string]interface{}{
				"enabled": true,
				"hosts": []map[string]interface{}{
					{
						"host": "headlamp.madhan.app",
						"paths": []map[string]interface{}{
							{"path": "/", "type": "ImplementationSpecific"},
						},
					},
				},
				"tls": []map[string]interface{}{
					{
						"secretName": "headlamp-tls",
						"hosts":      []interface{}{"headlamp.madhan.app"},
					},
				},
			},
		},
	})
	return chart
}
