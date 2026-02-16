package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/k8s"
)

func NewAlertManagerChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
		},
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

	// CRDs
	// CRDs
	cdk8s.NewInclude(chart, jsii.String("prometheus-crds"), &cdk8s.IncludeProps{
		Url: jsii.String("https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-82.0.1/charts/kube-prometheus-stack/charts/crds/crds/crd-prometheusrules.yaml"),
	})
	cdk8s.NewInclude(chart, jsii.String("servicemonitor-crds"), &cdk8s.IncludeProps{
		Url: jsii.String("https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-82.0.1/charts/kube-prometheus-stack/charts/crds/crds/crd-servicemonitors.yaml"),
	})
	cdk8s.NewInclude(chart, jsii.String("alertmanager-crds"), &cdk8s.IncludeProps{
		Url: jsii.String("https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-82.0.1/charts/kube-prometheus-stack/charts/crds/crds/crd-alertmanagers.yaml"),
	})
	cdk8s.NewInclude(chart, jsii.String("prometheus-crd"), &cdk8s.IncludeProps{
		Url: jsii.String("https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-82.0.1/charts/kube-prometheus-stack/charts/crds/crds/crd-prometheuses.yaml"),
	})

	cdk8s.NewHelm(chart, jsii.String("alertmanager-only"), &cdk8s.HelmProps{
		Chart:       jsii.String("kube-prometheus-stack"),
		Repo:        jsii.String("https://prometheus-community.github.io/helm-charts"),
		Version:     jsii.String("82.0.1"),
		ReleaseName: jsii.String("alertmanager"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
