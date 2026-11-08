package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/metricsserver"
)

// NewMetricsServerChart deploys Kubernetes Metrics Server into kube-system.
// The control-plane toleration only lets the server schedule there; it scrapes every
// kubelet wherever it runs.
func NewMetricsServerChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String("kube-system"),
	})

	values := map[string]any{
		"args": []string{
			// Talos kubelets serve self-signed certificates.
			"--kubelet-insecure-tls",
			// Node hostnames are not resolvable in cluster DNS.
			"--kubelet-preferred-address-types=InternalIP",
		},
		"tolerations": []map[string]any{
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
