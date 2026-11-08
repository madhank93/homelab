package observability

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/kit"
)

const (
	vmInsertEndpoint = "http://vmsingle-vm-stack.victoria-metrics.svc.cluster.local:8428/api/v1/write"
	vlOtlpEndpoint   = "http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/insert/opentelemetry"
)

// NewOtelCollectorChart deploys two OpenTelemetry Collector releases:
//
//   - Agent (DaemonSet): container logs, kubelet stats and host metrics from every node.
//   - Gateway (Deployment): cluster-level resource metrics and Kubernetes events.
//
// Both send metrics to VictoriaMetrics by Prometheus remote write and logs to
// VictoriaLogs over OTLP/HTTP.
func NewOtelCollectorChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// The agent DaemonSet uses hostPath volumes and hostPorts.
	kit.Namespace(chart, namespace, true)

	// The contrib image carries the Kubernetes receivers (kubeletstats, k8s_cluster, filelog).
	img := config.ContainerImage("otel-collector-contrib")
	image := map[string]any{"repository": img.Image, "tag": img.Tag}

	commonExporters := map[string]any{
		"prometheusremotewrite": map[string]any{
			"endpoint": vmInsertEndpoint,
			// Turns resource attributes such as k8s.pod.name into labels.
			"resource_to_telemetry_conversion": map[string]any{"enabled": true},
		},
		"otlp_http/logs": map[string]any{
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

	// Agent: presets add the filelog, kubeletstats and hostmetrics receivers, the
	// k8sattributes processor, and their RBAC and mounts.
	agentValues := map[string]any{
		"mode": "daemonset",
		"podAnnotations": map[string]any{
			"reloader.stakater.com/auto": "true",
		},
		"image": image,
		"presets": map[string]any{
			// Reads /var/log/pods on each node.
			"logsCollection": map[string]any{
				"enabled":              true,
				"includeCollectorLogs": false,
			},
			// The preset dials the kubelet by node name, which cluster DNS cannot
			// resolve; receivers.kubeletstats below dials the node IP instead.
			"kubeletMetrics": map[string]any{"enabled": true},
			"hostMetrics":    map[string]any{"enabled": true},
			"kubernetesAttributes": map[string]any{
				"enabled":             true,
				"extractAllPodLabels": true,
			},
		},
		"config": map[string]any{
			"extensions": map[string]any{
				// Keeps filelog read offsets on the host, so a restarted agent does
				// not resend logs it already shipped.
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
						"exporters":  []string{"otlp_http/logs"},
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
		"tolerations": []map[string]any{
			{"operator": "Exists"},
		},
		// Host directory for the file_storage checkpoints above.
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

	otel := config.HelmChart("opentelemetry-collector")
	cdk8s.NewHelm(chart, jsii.String("otel-agent"), &cdk8s.HelmProps{
		Chart:       jsii.String("opentelemetry-collector"),
		Repo:        jsii.String(otel.Repo),
		Version:     jsii.String(otel.Version),
		ReleaseName: jsii.String("otel-agent"),
		Namespace:   jsii.String(namespace),
		Values:      &agentValues,
	})

	// Gateway: presets add the k8s_cluster receiver, the k8sobjects receiver
	// (events as logs) and their RBAC.
	gatewayValues := map[string]any{
		"mode": "deployment",
		"podAnnotations": map[string]any{
			"reloader.stakater.com/auto": "true",
		},
		"image": image,
		"presets": map[string]any{
			"clusterMetrics":   map[string]any{"enabled": true},
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
				// The preset also puts this processor in the events pipeline, and
				// events carry no pod IP or UID; without passthrough they are dropped.
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
						"exporters":  []string{"otlp_http/logs"},
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
		Repo:        jsii.String(otel.Repo),
		Version:     jsii.String(otel.Version),
		ReleaseName: jsii.String("otel-gateway"),
		Namespace:   jsii.String(namespace),
		Values:      &gatewayValues,
	})

	return chart
}
