package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewVictoriaLogsChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{
		"server": map[string]any{
			"enabled": true,
			"persistentVolume": map[string]any{
				"enabled": true,
				"size":    "100Gi",
			},
			"retention": "30d",
		},
		"fluent-bit": map[string]any{
			"enabled": true,
		},
		"service": map[string]any{
			"type": "ClusterIP",
			"port": 9428,
		},
	}

	cdk8s.NewHelm(chart, jsii.String("victoria-logs-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("victoria-logs-single"),
		Repo:        jsii.String("https://victoriametrics.github.io/helm-charts"),
		Version:     jsii.String("0.11.26"),
		ReleaseName: jsii.String("victoria-logs"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
