package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/victorialogssingle"
)

func NewVictoriaLogsChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{})

	values := map[string]interface{}{
		"server": map[string]interface{}{
			"enabled": true,
			"persistentVolume": map[string]interface{}{
				"enabled": true,
				"size":    "100Gi",
			},
			"retention": "30d",
		},
		"fluent-bit": map[string]interface{}{
			"enabled": true,
		},
		"service": map[string]interface{}{
			"type": "ClusterIP",
			"port": 9428,
		},
	}

	victorialogssingle.NewVictorialogssingle(chart, jsii.String("victoria-logs-release"), &victorialogssingle.VictorialogssingleProps{
		ReleaseName: jsii.String("victoria-logs"),
		Values:      &values,
	})

	return chart
}
