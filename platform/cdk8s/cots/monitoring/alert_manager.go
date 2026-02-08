package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewAlertManagerChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{
		"prometheus": map[string]any{
			"enabled": false,
		},
		"grafana": map[string]any{
			"enabled": false,
		},
		"kubeStateMetrics": map[string]any{
			"enabled": false,
		},
		"nodeExporter": map[string]any{
			"enabled": false,
		},
		"prometheusOperator": map[string]any{
			"enabled": true, // Keep for CRDs
		},
		// Alertmanager configuration
		"alertmanager": map[string]any{
			"enabled": true,
			"alertmanagerSpec": map[string]any{
				"replicas": 1,
				"storage": map[string]any{
					"volumeClaimTemplate": map[string]any{
						"spec": map[string]any{
							"resources": map[string]any{
								"requests": map[string]any{
									"storage": "10Gi",
								},
							},
						},
					},
				},
			},
			"config": map[string]any{
				"global": map[string]any{
					"smtp_smarthost": "localhost:587",
					"smtp_from":      "alertmanager@example.com",
				},
				"route": map[string]any{
					"group_by":        []string{"alertname"},
					"group_wait":      "10s",
					"group_interval":  "10s",
					"repeat_interval": "1h",
					"receiver":        "web.hook",
				},
				"receivers": []map[string]any{
					{
						"name": "web.hook",
					},
				},
			},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("alertmanager-only"), &cdk8s.HelmProps{
		Chart:       jsii.String("kube-prometheus-stack"),
		Repo:        jsii.String("https://prometheus-community.github.io/helm-charts"),
		Version:     jsii.String("67.6.1"),
		ReleaseName: jsii.String("alertmanager"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
