package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/victoriametricscluster"
)

func NewVictoriaMetricsChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String("monitoring"),
	})

	values := map[string]interface{}{
		"server": map[string]interface{}{
			"enabled": true,
			"persistentVolume": map[string]interface{}{
				"enabled": true,
				"size":    "50Gi",
			},
			"retention": "30d",
		},
		"vmselect": map[string]interface{}{
			"enabled":      true,
			"replicaCount": 1,
			"resources": map[string]interface{}{
				"limits": map[string]interface{}{
					"memory": "1Gi",
					"cpu":    "500m",
				},
			},
		},
		"vminsert": map[string]interface{}{
			"enabled":      true,
			"replicaCount": 1,
		},
		"vmstorage": map[string]interface{}{
			"enabled":      true,
			"replicaCount": 1,
			"persistentVolume": map[string]interface{}{
				"enabled": true,
				"size":    "100Gi",
			},
		},
	}

	victoriametricscluster.NewVictoriametricscluster(chart, jsii.String("victoria-metrics-release"), &victoriametricscluster.VictoriametricsclusterProps{
		ReleaseName: jsii.String("victoria-metrics"),
		Values:      &values,
	})

	return chart
}
