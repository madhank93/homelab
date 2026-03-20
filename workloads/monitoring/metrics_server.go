package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/metricsserver"
)

// NewMetricsServerChart deploys Kubernetes Metrics Server into kube-system.
//
// Talos-specific flags:
//   - --kubelet-insecure-tls: Talos kubelets use self-signed TLS certs.
//   - --kubelet-preferred-address-types=InternalIP: avoids hostname resolution
//     issues common in Talos clusters where node FQDNs are not resolvable.
//
// A control-plane toleration is added so the server can also scrape control-plane
// node metrics (required for accurate HPA on control-plane workloads).
func NewMetricsServerChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String("kube-system"),
	})

	values := map[string]any{
		"args": []string{
			// Talos kubelets use self-signed TLS — skip cert verification.
			"--kubelet-insecure-tls",
			// Use node InternalIP to avoid hostname resolution issues on Talos.
			"--kubelet-preferred-address-types=InternalIP",
		},
		"tolerations": []map[string]any{
			// Allow running on control-plane nodes so control-plane metrics are collected.
			{"key": "node-role.kubernetes.io/control-plane", "operator": "Exists", "effect": "NoSchedule"},
		},
	}

	metricsserver.NewMetricsserver(chart, jsii.String("metrics-server-release"), &metricsserver.MetricsserverProps{
		ReleaseName: jsii.String("metrics-server"),
		Namespace:   jsii.String("kube-system"),
		Values:      &values,
	})

	return chart
}
