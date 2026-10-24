package storage

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/longhorn"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewLonghornChart deploys Longhorn distributed block storage into the given
// namespace.
//
// Each worker has a 120 Gi disk with ~100 Gi free, so the 300%
// storageOverProvisioningPercentage below allows ~300 Gi of claims per node.
// The namespace is labelled privileged because the Longhorn CSI driver and
// instance manager require elevated host access (device files, kernel modules).
func NewLonghornChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, true)

	values := map[string]any{
		"defaultSettings": map[string]any{
			"defaultReplicaCount":           3,
			"defaultDataPath":               "/var/lib/longhorn/", // Talos persistent path
			"createDefaultDiskLabeledNodes": false,                // Auto-provision default disk on ALL nodes (not just labelled ones)
			// 300% of a 120Gi disk leaves room for new 50Gi replicas (AI model
			// volumes); 200% was exhausted.
			"storageOverProvisioningPercentage": 300,
		},
		"persistence": map[string]any{
			"defaultClass":             true,
			"defaultClassReplicaCount": 3,
		},
		// The chart's internal NetworkPolicies (enforced by Cilium) only admit
		// Longhorn's own pods to longhorn-manager:9500; vmagent scrapes from
		// another namespace and needs its own peer entry.
		"networkPolicies": map[string]any{
			"metricsScrapeSources": []map[string]any{
				{
					"namespaceSelector": map[string]any{
						"matchLabels": map[string]any{"kubernetes.io/metadata.name": "victoria-metrics"},
					},
					"podSelector": map[string]any{
						"matchLabels": map[string]any{"app.kubernetes.io/name": "vmagent"},
					},
				},
			},
		},
		"metrics": map[string]any{
			"serviceMonitor": map[string]any{
				"enabled": true,
			},
		},
		// Disable pre-upgrade hook for initial installation (ArgoCD/GitOps best practice)
		"preUpgradeChecker": map[string]any{
			"jobEnabled": false,
		},
		// longhorn-manager must not tolerate the control-plane taint: a manager pod
		// registers its node as a Longhorn node, and every registered node needs an
		// engine image. Engine-image DaemonSets get no tolerations, so control planes
		// would leave EngineImages stuck "deploying" and block volume attachment.
		"longhornManager": map[string]any{
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "1000m", "memory": "1Gi"},
				"requests": map[string]any{"cpu": "200m", "memory": "256Mi"},
			},
		},
		"longhornDriver": map[string]any{
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
			"tolerations": []map[string]any{
				{"key": "node-role.kubernetes.io/control-plane", "operator": "Exists", "effect": "NoSchedule"},
			},
		},
		"longhornUI": map[string]any{
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
			"tolerations": []map[string]any{
				{"key": "node-role.kubernetes.io/control-plane", "operator": "Exists", "effect": "NoSchedule"},
			},
		},
	}

	kit.Route(chart, "longhorn", "longhorn", "longhorn-frontend", 80)

	longhorn.NewLonghorn(chart, jsii.String("longhorn-release"), &longhorn.LonghornProps{
		ReleaseName: jsii.String("longhorn"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
		// helm template otherwise checks the chart's kubeVersion against the local
		// Helm binary's built-in default; this matches the k8s@ import in cdk8s.yaml.
		HelmFlags: &[]*string{jsii.String("--kube-version=1.36.0")},
	})

	return chart
}
