package security

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/kyverno"
)

// NewKyvernoChart deploys the Kyverno policy engine in high-availability mode.
//
// Replica counts:
//   - admissionController: 3 (odd number required for leader election quorum)
//   - backgroundController: 2
//   - cleanupController:    2
//   - reportsController:    2
//
// A ServiceMonitor on port 8000 is enabled for VictoriaMetrics scraping,
// and a Grafana dashboard ConfigMap is created for policy observability.
func NewKyvernoChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{

		// High availability configuration
		"admissionController": map[string]any{
			"replicas": 3,
			"resources": map[string]any{
				"limits": map[string]any{
					"cpu":    "1000m",
					"memory": "512Mi",
				},
				"requests": map[string]any{
					"cpu":    "100m",
					"memory": "256Mi",
				},
			},
		},

		"backgroundController": map[string]any{
			"replicas": 2,
			"resources": map[string]any{
				"limits": map[string]any{
					"cpu":    "500m",
					"memory": "256Mi",
				},
				"requests": map[string]any{
					"cpu":    "50m",
					"memory": "128Mi",
				},
			},
		},

		"cleanupController": map[string]any{
			"replicas": 2,
			"resources": map[string]any{
				"limits": map[string]any{
					"cpu":    "500m",
					"memory": "256Mi",
				},
				"requests": map[string]any{
					"cpu":    "50m",
					"memory": "128Mi",
				},
			},
		},

		"reportsController": map[string]any{
			"replicas": 2,
			"resources": map[string]any{
				"limits": map[string]any{
					"cpu":    "500m",
					"memory": "256Mi",
				},
				"requests": map[string]any{
					"cpu":    "50m",
					"memory": "128Mi",
				},
			},
		},

		// Service account configuration
		"serviceAccount": map[string]any{
			"create": true,
			"name":   "kyverno",
		},

		// Security and RBAC
		"podSecurityContext": map[string]any{
			"runAsNonRoot": true,
			"runAsUser":    65534,
			"fsGroup":      65534,
		},

		// Node selector for system nodes (optional)
		"nodeSelector": map[string]any{
			"kubernetes.io/os": "linux",
		},

		// Monitoring and metrics
		"metricsService": map[string]any{
			"create": true,
			"type":   "ClusterIP",
			"port":   8000,
		},

		// Webhook configurations
		"webhooksCleanup": map[string]any{
			"enabled": true,
		},

		// Policy exception handling
		"policyExceptions": map[string]any{
			"enabled": true,
		},

		// Image verification with Cosign (optional)
		"imageVerification": map[string]any{
			"enabled": false, // Set to true if you want image verification
		},

		// Grafana dashboard integration — creates a ConfigMap with the Kyverno
		// dashboard JSON; Grafana's sidecar picks it up automatically.
		"grafana": map[string]any{
			"enabled": true,
		},

		// Global settings
		"global": map[string]any{
			"image": map[string]any{
				"pullPolicy": "IfNotPresent",
			},
		},
	}

	kyverno.NewKyverno(chart, jsii.String("kyverno-release"), &kyverno.KyvernoProps{
		ReleaseName: jsii.String("kyverno"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// ServiceMonitor — tells VMAgent to scrape Kyverno's metrics on port 8000.
	cdk8s.NewApiObject(chart, jsii.String("kyverno-servicemonitor"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("monitoring.coreos.com/v1"),
		Kind:       jsii.String("ServiceMonitor"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("kyverno"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"selector": map[string]any{
			"matchLabels": map[string]any{"app.kubernetes.io/name": "kyverno"},
		},
		"namespaceSelector": map[string]any{
			"matchNames": []string{namespace},
		},
		"endpoints": []map[string]any{
			{
				"port":     "metrics-port",
				"path":     "/metrics",
				"interval": "30s",
			},
		},
	}))

	return chart
}
