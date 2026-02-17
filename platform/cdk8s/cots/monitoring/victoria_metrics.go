package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewVictoriaMetricsChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{
		"server": map[string]any{
			"enabled": true,
			"persistentVolume": map[string]any{
				"enabled": true,
				"size":    "50Gi",
			},
			"retention": "30d",
		},
		"vmselect": map[string]any{
			"enabled":      true,
			"replicaCount": 1,
			"resources": map[string]any{
				"limits": map[string]any{
					"memory": "1Gi",
					"cpu":    "500m",
				},
			},
		},
		"vminsert": map[string]any{
			"enabled":      true,
			"replicaCount": 1,
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
		},
		"vmstorage": map[string]any{
			"enabled":      true,
			"replicaCount": 1,
			"persistentVolume": map[string]any{
				"enabled": true,
				"size":    "100Gi",
			},
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "1000m", "memory": "1Gi"},
				"requests": map[string]any{"cpu": "200m", "memory": "256Mi"},
			},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("victoria-metrics-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("victoria-metrics-cluster"),
		Repo:        jsii.String("https://victoriametrics.github.io/helm-charts"),
		Version:     jsii.String("0.34.0"),
		ReleaseName: jsii.String("victoria-metrics"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
