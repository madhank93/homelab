package security

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/kyverno"
)

// NewKyvernoChart deploys the Kyverno policy engine in high-availability mode.
// The admission controller runs 3 replicas (odd, for leader-election quorum);
// the other controllers run 2. Metrics on port 8000 are scraped by VMAgent.
func NewKyvernoChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	controller := func(replicas int, limitCPU, limitMem, reqCPU, reqMem string) map[string]any {
		return map[string]any{
			"replicas": replicas,
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": limitCPU, "memory": limitMem},
				"requests": map[string]any{"cpu": reqCPU, "memory": reqMem},
			},
		}
	}

	values := map[string]any{
		"admissionController":  controller(3, "1000m", "512Mi", "100m", "256Mi"),
		"backgroundController": controller(2, "500m", "256Mi", "50m", "128Mi"),
		"cleanupController":    controller(2, "500m", "256Mi", "50m", "128Mi"),
		"reportsController":    controller(2, "500m", "256Mi", "50m", "128Mi"),
		"serviceAccount": map[string]any{
			"create": true,
			"name":   "kyverno",
		},
		"podSecurityContext": map[string]any{
			"runAsNonRoot": true,
			"runAsUser":    65534,
			"fsGroup":      65534,
		},
		"nodeSelector": map[string]any{
			"kubernetes.io/os": "linux",
		},
		"metricsService": map[string]any{
			"create": true,
			"type":   "ClusterIP",
			"port":   8000,
		},
		// The post-upgrade CRD migration hook fails under Argo CD: its ServiceAccount
		// is pruned before the Job runs (kyverno/kyverno#17741). Re-enable it, or run
		// the migration by hand, for any chart bump that changes a CRD storage version.
		"crds": map[string]any{
			"migration": map[string]any{"enabled": false},
		},
		"webhooksCleanup":   map[string]any{"enabled": true},
		"policyExceptions":  map[string]any{"enabled": true},
		"imageVerification": map[string]any{"enabled": false},
		// Renders the Kyverno dashboard ConfigMap for Grafana's sidecar.
		"grafana": map[string]any{"enabled": true},
		"global": map[string]any{
			"image": map[string]any{"pullPolicy": "IfNotPresent"},
		},
	}

	kyverno.NewKyverno(chart, jsii.String("kyverno-release"), &kyverno.KyvernoProps{
		ReleaseName: jsii.String("kyverno"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	cdk8s.NewApiObject(chart, jsii.String("kyverno-servicemonitor"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("monitoring.coreos.com/v1"),
		Kind:       jsii.String("ServiceMonitor"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String("kyverno")},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"selector": map[string]any{
			"matchLabels": map[string]any{"app.kubernetes.io/name": "kyverno"},
		},
		"namespaceSelector": map[string]any{
			"matchNames": []string{namespace},
		},
		"endpoints": []map[string]any{
			{"port": "metrics-port", "path": "/metrics", "interval": "30s"},
		},
	}))

	return chart
}
