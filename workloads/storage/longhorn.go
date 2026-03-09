package storage

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/imports/longhorn"
)

func NewLonghornChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Privileged namespace — Longhorn CSI driver and manager require elevated privileges.
	k8s.NewKubeNamespace(chart, jsii.String("longhorn-namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
			Labels: &map[string]*string{
				"pod-security.kubernetes.io/enforce": jsii.String("privileged"),
				"pod-security.kubernetes.io/audit":   jsii.String("privileged"),
				"pod-security.kubernetes.io/warn":    jsii.String("privileged"),
			},
		},
	})

	values := map[string]any{
		"defaultSettings": map[string]any{
			"defaultReplicaCount":           3,
			"defaultDataPath":               "/var/lib/longhorn/", // Talos persistent path
			"createDefaultDiskLabeledNodes": false,                // Auto-provision default disk on ALL nodes (not just labelled ones)
			// 200% overprovisioning: allows up to 240Gi scheduled per 120Gi disk.
			// Each worker node has ~100Gi actually free but only 5Gi scheduling
			// headroom at the default 100% limit. At 200%, headroom is ~125Gi per
			// node — enough to schedule 100Gi replicas for AI workloads.
			"storageOverProvisioningPercentage": 200,
		},
		"persistence": map[string]any{
			"defaultClass":             true,
			"defaultClassReplicaCount": 3,
		},
		// Disable pre-upgrade hook for initial installation (ArgoCD/GitOps best practice)
		"preUpgradeChecker": map[string]any{
			"jobEnabled": false,
		},
		// Talos-specific: Allow control-plane components to run on control-plane nodes
		"longhornManager": map[string]any{
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "1000m", "memory": "1Gi"},
				"requests": map[string]any{"cpu": "200m", "memory": "256Mi"},
			},
			"tolerations": []map[string]any{
				{"key": "node-role.kubernetes.io/control-plane", "operator": "Exists", "effect": "NoSchedule"},
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

	longhorn.NewLonghorn(chart, jsii.String("longhorn-release"), &longhorn.LonghornProps{
		ReleaseName: jsii.String("longhorn"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
