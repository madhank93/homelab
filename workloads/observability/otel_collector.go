package observability

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

const (
	vmInsertEndpoint = "http://victoria-metrics-victoria-metrics-cluster-vminsert.victoria-metrics.svc.cluster.local:8480/insert/0/prometheus/api/v1/write"
	// VictoriaLogs OTLP endpoint — collector appends /v1/logs automatically
	vlOtlpEndpoint = "http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/insert/opentelemetry"
)

// NewOtelCollectorChart deploys two OpenTelemetry Collector instances:
//
//   - Agent (DaemonSet): runs on every node, collects container logs,
//     node/pod/container metrics from kubelet, and host-level metrics.
//
//   - Gateway (Deployment): cluster-scoped receiver that collects
//     Kubernetes resource metrics (nodes, pods, deployments) and
//     Kubernetes events.
//
// Both export metrics to VictoriaMetrics (prometheus remote write) and
// logs to VictoriaLogs (OTLP/HTTP).
func NewOtelCollectorChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
			Labels: &map[string]*string{
				// Agent DaemonSet uses hostPath volumes and hostPorts — requires privileged
				"pod-security.kubernetes.io/enforce": jsii.String("privileged"),
			},
		},
	})

	commonExporters := map[string]any{
		"prometheusremotewrite": map[string]any{
			"endpoint": vmInsertEndpoint,
			// Converts resource attributes (k8s.pod.name etc.) to Prometheus labels
			"resource_to_telemetry_conversion": map[string]any{"enabled": true},
		},
		"otlphttp/logs": map[string]any{
			"endpoint": vlOtlpEndpoint,
			"tls":      map[string]any{"insecure": true},
		},
	}

	commonProcessors := map[string]any{
		"batch": map[string]any{"timeout": "10s"},
		"memory_limiter": map[string]any{
			"check_interval":         "5s",
			"limit_percentage":       80,
			"spike_limit_percentage": 25,
		},
	}

	// ── Agent (DaemonSet) ────────────────────────────────────────────────────
	// Presets configure: filelog receiver, kubeletstats receiver, hostmetrics
	// receiver, k8sattributes processor, and the necessary RBAC + volume mounts.
	agentValues := map[string]any{
		"mode": "daemonset",
		// contrib image includes all Kubernetes receivers (kubeletstats, k8s_cluster, filelog, etc.)
		"image": map[string]any{"repository": "otel/opentelemetry-collector-contrib"},
		"presets": map[string]any{
			// Container log collection from /var/log/pods on each node
			"logsCollection": map[string]any{
				"enabled":              true,
				"includeCollectorLogs": false,
			},
			// Node, pod, container, volume metrics from kubelet /stats/summary
			"kubeletMetrics": map[string]any{"enabled": true},
			// CPU, memory, disk, network metrics from the host OS
			"hostMetrics": map[string]any{"enabled": true},
			// Enriches telemetry with k8s.pod.name, k8s.namespace.name, etc.
			"kubernetesAttributes": map[string]any{
				"enabled":             true,
				"extractAllPodLabels": true,
			},
		},
		"config": map[string]any{
			"exporters":  commonExporters,
			"processors": commonProcessors,
			"service": map[string]any{
				"pipelines": map[string]any{
					"logs": map[string]any{
						"receivers":  []string{"filelog"},
						"processors": []string{"memory_limiter", "k8sattributes", "batch"},
						"exporters":  []string{"otlphttp/logs"},
					},
					"metrics": map[string]any{
						"receivers":  []string{"kubeletstats", "hostmetrics"},
						"processors": []string{"memory_limiter", "k8sattributes", "batch"},
						"exporters":  []string{"prometheusremotewrite"},
					},
				},
			},
		},
		"resources": map[string]any{
			"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
			"requests": map[string]any{"cpu": "50m", "memory": "128Mi"},
		},
		// Run on every node including control plane
		"tolerations": []map[string]any{
			{"operator": "Exists"},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("otel-agent"), &cdk8s.HelmProps{
		Chart:       jsii.String("opentelemetry-collector"),
		Repo:        jsii.String("https://open-telemetry.github.io/opentelemetry-helm-charts"),
		Version:     jsii.String("0.108.0"),
		ReleaseName: jsii.String("otel-agent"),
		Namespace:   jsii.String(namespace),
		Values:      &agentValues,
	})

	// ── Gateway (Deployment) ─────────────────────────────────────────────────
	// Presets configure: k8s_cluster receiver (node/pod/deployment metrics),
	// k8sobjects receiver (Kubernetes events as logs), and necessary RBAC.
	gatewayValues := map[string]any{
		"mode":  "deployment",
		"image": map[string]any{"repository": "otel/opentelemetry-collector-contrib"},
		"presets": map[string]any{
			// Cluster-level resource metrics: nodes, pods, deployments, etc.
			"clusterMetrics": map[string]any{"enabled": true},
			// Kubernetes events forwarded as log records to VictoriaLogs
			"kubernetesEvents": map[string]any{"enabled": true},
			"kubernetesAttributes": map[string]any{
				"enabled": true,
			},
		},
		"config": map[string]any{
			"exporters":  commonExporters,
			"processors": commonProcessors,
			"service": map[string]any{
				"pipelines": map[string]any{
					"metrics": map[string]any{
						"receivers":  []string{"k8s_cluster"},
						"processors": []string{"memory_limiter", "k8sattributes", "batch"},
						"exporters":  []string{"prometheusremotewrite"},
					},
					"logs": map[string]any{
						"receivers":  []string{"k8sobjects"},
						"processors": []string{"memory_limiter", "batch"},
						"exporters":  []string{"otlphttp/logs"},
					},
				},
			},
		},
		"replicaCount": 1,
		"resources": map[string]any{
			"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
			"requests": map[string]any{"cpu": "50m", "memory": "128Mi"},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("otel-gateway"), &cdk8s.HelmProps{
		Chart:       jsii.String("opentelemetry-collector"),
		Repo:        jsii.String("https://open-telemetry.github.io/opentelemetry-helm-charts"),
		Version:     jsii.String("0.108.0"),
		ReleaseName: jsii.String("otel-gateway"),
		Namespace:   jsii.String(namespace),
		Values:      &gatewayValues,
	})

	return chart
}
