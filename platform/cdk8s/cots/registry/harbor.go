package registry

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/harbor"
)

func NewHarborChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{})

	values := map[string]interface{}{
		"expose": map[string]interface{}{
			"type": "ingress",
			"ingress": map[string]interface{}{
				"hosts": map[string]interface{}{
					"core": "harbor.local",
				},
			},
		},
		"externalURL": "https://harbor.local",
		"persistence": map[string]interface{}{
			"enabled": true,
			"persistentVolumeClaim": map[string]interface{}{
				"registry": map[string]interface{}{
					"size": "50Gi",
				},
				"database": map[string]interface{}{
					"size": "10Gi",
				},
			},
		},
		"harborAdminPassword": "Harbor12345",
	}

	harbor.NewHarbor(chart, jsii.String("harbor-release"), &harbor.HarborProps{
		ReleaseName: jsii.String("harbor"),
		Values:      &values,
	})

	return chart
}
