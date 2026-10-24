package observability

import (
	"strings"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewVictoriaMetricsChart deploys the victoria-metrics-k8s-stack chart:
// VMOperator, VMSingle (not VMCluster), VMAgent, VMAlert, Alertmanager,
// node-exporter, kube-state-metrics and the default dashboards and rules.
func NewVictoriaMetricsChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// node-exporter needs hostNetwork, hostPID, hostPath and hostPort 9100.
	kit.Namespace(chart, namespace, true)

	vmStack := config.HelmChart("victoria-metrics-k8s-stack")
	vmOperator := config.HelmChart("victoria-metrics-operator").Version
	promStack := config.HelmChart("kube-prometheus-stack").Version

	// cdk8s.NewHelm never renders a chart's crds/ directory, so CRDs are
	// included explicitly. The victoria-metrics-operator pin must match the
	// operator version in the stack chart's Chart.lock.
	cdk8s.NewInclude(chart, jsii.String("vm-operator-crds"), &cdk8s.IncludeProps{
		Url: jsii.String("https://raw.githubusercontent.com/VictoriaMetrics/helm-charts/victoria-metrics-operator-" + vmOperator + "/charts/victoria-metrics-operator/crd.yaml"),
	})

	// ServiceMonitor and PrometheusRule CRDs: many apps ship ServiceMonitors,
	// and VMOperator converts them.
	promCRDs := "https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-" + promStack + "/charts/kube-prometheus-stack/charts/crds/crds/"
	cdk8s.NewInclude(chart, jsii.String("servicemonitor-crd"), &cdk8s.IncludeProps{
		Url: jsii.String(promCRDs + "crd-servicemonitors.yaml"),
	})
	cdk8s.NewInclude(chart, jsii.String("prometheusrule-crd"), &cdk8s.IncludeProps{
		Url: jsii.String(promCRDs + "crd-prometheusrules.yaml"),
	})

	cdk8s.NewHelm(chart, jsii.String("victoria-metrics-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("victoria-metrics-k8s-stack"),
		Repo:        jsii.String(vmStack.Repo),
		Version:     jsii.String(vmStack.Version),
		ReleaseName: jsii.String("victoria-metrics"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
			// The default 44-char fullname pushes VMAlertmanager pod labels past
			// the 63-byte limit. Services become vmsingle-vm-stack, etc.
			"fullnameOverride": "vm-stack",

			"vmsingle": map[string]any{
				"enabled": true,
				"spec": map[string]any{
					"retentionPeriod": "30d",
					"storage": map[string]any{
						"volumeClaimTemplate": map[string]any{
							"spec": map[string]any{
								"resources": map[string]any{
									"requests": map[string]any{"storage": "100Gi"},
								},
							},
						},
					},
					"resources": map[string]any{
						"limits":   map[string]any{"cpu": "1000m", "memory": "2Gi"},
						"requests": map[string]any{"cpu": "200m", "memory": "512Mi"},
					},
				},
			},
			"vmcluster": map[string]any{"enabled": false},

			// cdk8s keeps helm.sh/hook annotations, so Argo CD would run this
			// cleanup hook as a pre-delete hook on every sync and fail.
			"victoria-metrics-operator": map[string]any{
				"crds": map[string]any{
					"cleanup": map[string]any{"enabled": false},
				},
			},

			// selectAllByDefault picks up every ServiceMonitor/PodMonitor cluster-wide.
			"vmagent": map[string]any{
				"enabled": true,
				"spec": map[string]any{
					"selectAllByDefault": true,
					"scrapeInterval":     "30s",
					"resources": map[string]any{
						"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
						"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
					},
					"tolerations": []map[string]any{
						{"operator": "Exists"},
					},
					"inlineScrapeConfig": `- job_name: "openbao"
  static_configs:
    - targets: ["` + strings.TrimPrefix(config.C.OpenBaoAddr, "http://") + `"]
  metrics_path: /v1/sys/metrics
  params:
    format: [prometheus]
`,
				},
			},

			"vmalert": map[string]any{
				"enabled": true,
				"spec": map[string]any{
					"resources": map[string]any{
						"limits":   map[string]any{"cpu": "200m", "memory": "256Mi"},
						"requests": map[string]any{"cpu": "50m", "memory": "64Mi"},
					},
				},
			},

			// No notification channels yet: every alert goes to the blackhole receiver.
			"alertmanager": map[string]any{
				"enabled": true,
				"spec": map[string]any{
					"replicaCount": 1,
					"resources": map[string]any{
						"limits":   map[string]any{"cpu": "200m", "memory": "256Mi"},
						"requests": map[string]any{"cpu": "50m", "memory": "64Mi"},
					},
				},
				"config": map[string]any{
					"route": map[string]any{
						"group_by":        []string{"alertname", "namespace"},
						"group_wait":      "10s",
						"group_interval":  "10m",
						"repeat_interval": "1h",
						"receiver":        "blackhole",
					},
					"receivers": []map[string]any{
						{"name": "blackhole"},
					},
				},
			},

			// Grafana disabled — deployed separately in monitoring/grafana.go.
			"grafana": map[string]any{"enabled": false},

			"prometheus-node-exporter": map[string]any{
				"enabled": true,
				"tolerations": []map[string]any{
					{"operator": "Exists"},
				},
			},

			"kube-state-metrics": map[string]any{"enabled": true},

			"kubelet":               map[string]any{"enabled": true},
			"kubeApiServer":         map[string]any{"enabled": true},
			"coreDns":               map[string]any{"enabled": true},
			"kubeEtcd":              map[string]any{"enabled": true},
			"kubeScheduler":         map[string]any{"enabled": true},
			"kubeControllerManager": map[string]any{"enabled": true},
			// Cilium replaces kube-proxy — no kube-proxy metrics endpoint
			"kubeProxy": map[string]any{"enabled": false},

			// Dashboards and rules are not rendered by the chart: its sync-job (an
			// Argo CD PostSync hook) fetches them from upstream GitHub and creates
			// the ConfigMaps and VMRules, owned by its ServiceAccount.
			"defaultDashboards": map[string]any{"enabled": true},
			"defaultRules":      map[string]any{"enabled": true},
		},
	})

	// VMSingle serves its UI at /vmui/, so the bare root redirects there.
	kit.RouteRules(chart, "victoria-metrics", "vmselect", []map[string]any{
		{
			"matches": []map[string]any{
				{"path": map[string]any{"type": "Exact", "value": "/"}},
			},
			"filters": []map[string]any{
				{
					"type": "RequestRedirect",
					"requestRedirect": map[string]any{
						"path":       map[string]any{"type": "ReplaceFullPath", "replaceFullPath": "/vmui/"},
						"statusCode": 302,
					},
				},
			},
		},
		{
			"matches": []map[string]any{
				{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
			},
			"backendRefs": []map[string]any{kit.Backend("vmsingle-vm-stack", 8428)},
		},
	})

	kit.Route(chart, "alertmanager", "alertmanager", "vmalertmanager-vm-stack", 9093)

	return chart
}
