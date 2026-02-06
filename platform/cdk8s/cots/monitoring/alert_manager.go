package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/kubeprometheusstack"
)

func NewAlertManagerChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{
		"prometheus": map[string]interface{}{
			"enabled": false,
		},
		"grafana": map[string]interface{}{
			"enabled": false,
		},
		"kubeStateMetrics": map[string]interface{}{
			"enabled": false,
		},
		"nodeExporter": map[string]interface{}{
			"enabled": false,
		},
		"prometheusOperator": map[string]interface{}{
			"enabled": true, // Keep for CRDs
		},
		// Alertmanager configuration
		"alertmanager": map[string]interface{}{
			"enabled": true,
			"alertmanagerSpec": map[string]interface{}{
				"replicas": 1,
				"storage": map[string]interface{}{
					"volumeClaimTemplate": map[string]interface{}{
						"spec": map[string]interface{}{
							"resources": map[string]interface{}{
								"requests": map[string]interface{}{
									"storage": "10Gi",
								},
							},
						},
					},
				},
			},
			"config": map[string]interface{}{
				"global": map[string]interface{}{
					"smtp_smarthost": "localhost:587",
					"smtp_from":      "alertmanager@example.com",
				},
				"route": map[string]interface{}{
					"group_by":        []string{"alertname"},
					"group_wait":      "10s",
					"group_interval":  "10s",
					"repeat_interval": "1h",
					"receiver":        "web.hook",
				},
				"receivers": []map[string]interface{}{
					{
						"name": "web.hook",
					},
				},
			},
		},
	}

	// Use the actual constructor function name from your generated code
	kubeprometheusstack.NewKubeprometheusstack(chart, jsii.String("alertmanager-only"), &kubeprometheusstack.KubeprometheusstackProps{
		ReleaseName: jsii.String("alertmanager"),
		Values:      &values,
	})

	return chart
}
