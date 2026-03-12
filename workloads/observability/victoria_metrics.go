package observability

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

func NewVictoriaMetricsChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("victoria-metrics-namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
		},
	})

	// CRDs — cdk8s.NewHelm renders only Helm templates, never the chart's crds/ directory.
	// All CRDs that other resources depend on must be included explicitly here.

	// VMOperator CRDs (VMSingle, VMAgent, VMAlert, VMAlertmanager, VMRule, VMServiceScrape, etc.)
	// Pinned to the exact operator version bundled in victoria-metrics-k8s-stack@0.72.4 (Chart.lock: operator@0.59.2)
	cdk8s.NewInclude(chart, jsii.String("vm-operator-crds"), &cdk8s.IncludeProps{
		Url: jsii.String("https://raw.githubusercontent.com/VictoriaMetrics/helm-charts/victoria-metrics-operator-0.59.2/charts/victoria-metrics-operator/crd.yaml"),
	})

	// Prometheus-operator CRDs (ServiceMonitor, PrometheusRule) — kept for compatibility.
	// VMOperator watches ServiceMonitor CRDs when they exist. Apps like ArgoCD monitor,
	// Falco, Longhorn, and DCGM Exporter create ServiceMonitor resources.
	// Previously installed by alert_manager.go (kube-prometheus-stack); moved here to
	// prevent ArgoCD from pruning them when that app was removed.
	cdk8s.NewInclude(chart, jsii.String("servicemonitor-crd"), &cdk8s.IncludeProps{
		Url: jsii.String("https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-82.0.1/charts/kube-prometheus-stack/charts/crds/crds/crd-servicemonitors.yaml"),
	})
	cdk8s.NewInclude(chart, jsii.String("prometheusrule-crd"), &cdk8s.IncludeProps{
		Url: jsii.String("https://raw.githubusercontent.com/prometheus-community/helm-charts/kube-prometheus-stack-82.0.1/charts/kube-prometheus-stack/charts/crds/crds/crd-prometheusrules.yaml"),
	})

	// victoria-metrics-k8s-stack — all-in-one chart that includes:
	//   VMOperator  — watches ServiceMonitor/PodMonitor CRDs cluster-wide (replaces standalone vmagent)
	//   VMSingle    — single-node storage (replaces victoria-metrics-cluster with 1 replica each)
	//   VMAgent     — scrapes all ServiceMonitors automatically via selectAllByDefault
	//   VMAlert     — evaluates PrometheusRules / VMRules
	//   VMAlertmanager — replaces kube-prometheus-stack alertmanager
	//   node-exporter   — node CPU/mem/disk/network metrics (DaemonSet)
	//   kube-state-metrics — Kubernetes object metrics
	//   default dashboards — ConfigMaps with grafana_dashboard:"1" → Grafana auto-provisions
	cdk8s.NewHelm(chart, jsii.String("victoria-metrics-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("victoria-metrics-k8s-stack"),
		Repo:        jsii.String("https://victoriametrics.github.io/helm-charts"),
		Version:     jsii.String("0.72.4"),
		ReleaseName: jsii.String("victoria-metrics"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
			// Shorten generated resource names — the release name "victoria-metrics" combined
			// with chart name "victoria-metrics-k8s-stack" produces 44-char fullnames.
			// VMAlertmanager pod labels then exceed Kubernetes' 63-byte limit.
			// fullnameOverride: "vm-stack" gives a 8-char base → all labels stay short.
			// Service names become: vmsingle-vm-stack, vmalertmanager-vm-stack, etc.
			"fullnameOverride": "vm-stack",

			// VMSingle — single-node storage, simpler than VMCluster for homelab.
			// Service: vmsingle-vm-stack.victoria-metrics.svc.cluster.local:8429
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
			// VMCluster disabled — VMSingle is sufficient for single-node homelab.
			"vmcluster": map[string]any{"enabled": false},

			// VMAgent — scrapes all ServiceMonitors/PodMonitors cluster-wide.
			// selectAllByDefault:true picks up existing ServiceMonitors from ArgoCD, Falco,
			// Longhorn, DCGM Exporter, Kyverno, CNPG without any manual configuration.
			// Static scrape configs for services without ServiceMonitors (OpenBao).
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
					// Static scrape for OpenBao — custom metrics path not available via ServiceMonitor.
					"inlineScrapeConfig": `- job_name: "openbao"
  static_configs:
    - targets: ["openbao.openbao.svc.cluster.local:8200"]
  metrics_path: /v1/sys/metrics
  params:
    format: [prometheus]
`,
				},
			},

			// VMAlert — evaluates PrometheusRules and VMRules against VMSingle.
			"vmalert": map[string]any{
				"enabled": true,
				"spec": map[string]any{
					"resources": map[string]any{
						"limits":   map[string]any{"cpu": "200m", "memory": "256Mi"},
						"requests": map[string]any{"cpu": "50m", "memory": "64Mi"},
					},
				},
			},

			// Alertmanager — replaces kube-prometheus-stack (alertmanager-only).
			// Basic no-op config; add notification channels (Slack, PagerDuty, etc.) when needed.
			// Service: vmalertmanager-victoria-metrics.victoria-metrics.svc.cluster.local:9093
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

			// Node exporter — per-node CPU/memory/disk/network metrics (DaemonSet).
			// Tolerates all taints so it runs on control planes + GPU node.
			"prometheus-node-exporter": map[string]any{
				"enabled": true,
				"tolerations": []map[string]any{
					{"operator": "Exists"},
				},
			},

			// Kubernetes object metrics (Deployments, Pods, PVCs, etc.)
			"kube-state-metrics": map[string]any{"enabled": true},

			// Kubernetes component monitoring
			"kubelet":               map[string]any{"enabled": true},
			"kubeApiServer":         map[string]any{"enabled": true},
			"coreDns":               map[string]any{"enabled": true},
			"kubeEtcd":              map[string]any{"enabled": true},
			"kubeScheduler":         map[string]any{"enabled": true},
			"kubeControllerManager": map[string]any{"enabled": true},
			// Cilium replaces kube-proxy — no kube-proxy metrics endpoint
			"kubeProxy": map[string]any{"enabled": false},

			// Default dashboards as ConfigMaps labelled grafana_dashboard:"1".
			// Grafana sidecar auto-provisions: node-exporter, VM operator, VMAlert, k8s dashboards.
			"defaultDashboards": map[string]any{"enabled": true},

			// Default alerting rules for k8s + VictoriaMetrics components.
			"defaultRules": map[string]any{"create": true},
		},
	})

	// Gateway API HTTPRoute — vmselect.madhan.app → vmsingle-victoria-metrics:8429
	// VMSingle exposes /vmui/ directly (no /select/0/ prefix unlike VMCluster).
	cdk8s.NewApiObject(chart, jsii.String("victoria-metrics-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("victoria-metrics"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"vmselect.madhan.app"},
		"rules": []map[string]any{
			{
				// Redirect bare root to the vmui path (VMSingle uses /vmui/ not /select/0/vmui/)
				"matches": []map[string]any{
					{"path": map[string]any{"type": "Exact", "value": "/"}},
				},
				"filters": []map[string]any{
					{
						"type": "RequestRedirect",
						"requestRedirect": map[string]any{
							"path": map[string]any{
								"type":            "ReplaceFullPath",
								"replaceFullPath": "/vmui/",
							},
							"statusCode": 302,
						},
					},
				},
			},
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "vmsingle-vm-stack", "port": 8429, "weight": 1},
				},
			},
		},
	}))

	// Gateway API HTTPRoute — alertmanager.madhan.app → vmalertmanager:9093
	// Consolidated from alert_manager.go (was in separate alertmanager namespace).
	cdk8s.NewApiObject(chart, jsii.String("alertmanager-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("alertmanager"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"alertmanager.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "vmalertmanager-vm-stack", "port": 9093, "weight": 1},
				},
			},
		},
	}))

	return chart
}
