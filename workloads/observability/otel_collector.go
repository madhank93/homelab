package observability

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

const (
	// vmInsertEndpoint is the VictoriaMetrics single-node remote-write endpoint.
	// VMCluster is disabled (vmcluster.enabled=false in victoria_metrics.go).
	vmInsertEndpoint = "http://vmsingle-vm-stack.victoria-metrics.svc.cluster.local:8428/api/v1/write"
	// vlOtlpEndpoint is the VictoriaLogs OTLP/HTTP endpoint.
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
		"podAnnotations": map[string]any{
			"reloader.stakater.com/auto": "true",
		},
		// contrib image includes all Kubernetes receivers (kubeletstats, k8s_cluster, filelog, etc.)
		"image": map[string]any{"repository": "otel/opentelemetry-collector-contrib"},
		"presets": map[string]any{
			// Container log collection from /var/log/pods on each node
			"logsCollection": map[string]any{
				"enabled":              true,
				"includeCollectorLogs": false,
			},
			// Node, pod, container, volume metrics from kubelet /stats/summary.
			// The preset sets endpoint to "${K8S_NODE_NAME}:10250" — on Talos, node
			// hostnames are not in cluster DNS. We override the endpoint below using
			// K8S_NODE_IP (Downward API) so the agent connects via IP instead.
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
			"extensions": map[string]any{
				// file_storage persists filelog receiver read offsets (checkpoints)
				// to the hostPath volume so that agent restarts don't re-read historical
				// logs from the beginning, avoiding duplicates and slow recovery.
				"file_storage": map[string]any{
					"directory": "/var/lib/otelcol",
				},
			},
			"receivers": map[string]any{
				"kubeletstats": map[string]any{
					"endpoint":             "https://${env:K8S_NODE_IP}:10250",
					"insecure_skip_verify": true,
				},
			},
			"exporters":  commonExporters,
			"processors": commonProcessors,
			"service": map[string]any{
				"extensions": []string{"health_check", "file_storage"},
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
		// Persist filelog receiver checkpoints on the host so that agent restarts
		// (e.g. on config changes) do not re-read logs from the beginning,
		// which would cause duplicate entries and slow startup.
		"extraVolumes": []map[string]any{
			{
				"name": "otelcol-checkpoint",
				"hostPath": map[string]any{
					"path": "/var/lib/otelcol",
					"type": "DirectoryOrCreate",
				},
			},
		},
		"extraVolumeMounts": []map[string]any{
			{
				"name":      "otelcol-checkpoint",
				"mountPath": "/var/lib/otelcol",
			},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("otel-agent"), &cdk8s.HelmProps{
		Chart:       jsii.String("opentelemetry-collector"),
		Repo:        jsii.String("https://open-telemetry.github.io/opentelemetry-helm-charts"),
		Version:     jsii.String("0.156.2"),
		ReleaseName: jsii.String("otel-agent"),
		Namespace:   jsii.String(namespace),
		Values:      &agentValues,
	})

	// ── Gateway (Deployment) ─────────────────────────────────────────────────
	// Presets configure: k8s_cluster receiver (node/pod/deployment metrics),
	// k8sobjects receiver (Kubernetes events as logs), and necessary RBAC.
	gatewayValues := map[string]any{
		"mode":  "deployment",
		"podAnnotations": map[string]any{
			"reloader.stakater.com/auto": "true",
		},
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
			"exporters": commonExporters,
			"processors": map[string]any{
				"batch":          map[string]any{"timeout": "10s"},
				"memory_limiter": commonProcessors["memory_limiter"],
				// passthrough: true — k8sobjects log records (k8s Events) have no pod
				// IP or UID, so the default passthrough=false silently drops them all
				// before they reach the VictoriaLogs exporter.
				"k8sattributes": map[string]any{"passthrough": true},
			},
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
		Version:     jsii.String("0.156.2"),
		ReleaseName: jsii.String("otel-gateway"),
		Namespace:   jsii.String(namespace),
		Values:      &gatewayValues,
	})

	return chart
}
