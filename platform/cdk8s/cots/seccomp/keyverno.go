package seccomp

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

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

		// Grafana dashboard integration
		"grafana": map[string]any{
			"enabled": false, // Set to true if you have Grafana
		},

		// Global settings
		"global": map[string]any{
			"image": map[string]any{
				"pullPolicy": "IfNotPresent",
			},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("kyverno-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("kyverno"),
		Repo:        jsii.String("https://kyverno.github.io/kyverno"),
		Version:     jsii.String("3.7.0"),
		ReleaseName: jsii.String("kyverno"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
