package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/grafana"
)

func NewGrafanaChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String("monitoring"),
	})

	values := map[string]interface{}{
		"adminPassword": "admin123",
		"persistence": map[string]interface{}{
			"enabled": true,
			"size":    "10Gi",
		},
		"service": map[string]interface{}{
			"type": "ClusterIP",
			"port": 3000,
		},
		"ingress": map[string]interface{}{
			"enabled": true,
			"hosts":   []string{"grafana.local"},
		},
	}

	grafana.NewGrafana(chart, jsii.String("grafana-release"), &grafana.GrafanaProps{
		ReleaseName: jsii.String("grafana"),
		Values:      &values,
	})

	return chart
}
