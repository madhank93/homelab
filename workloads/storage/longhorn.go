package storage

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/imports/longhorn"
)

// NewLonghornChart deploys Longhorn distributed block storage (v1.10.2) into the
// given namespace.
//
// Storage layout per worker node: 120 Gi disk, ~100 Gi free.
// overProvisioningPercentage=200 gives ~240 Gi of schedulable storage per node.
// The namespace is labelled privileged because the Longhorn CSI driver and
// instance manager require elevated host access (device files, kernel modules).
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
			// 300% overprovisioning: allows up to 360Gi scheduled per 120Gi disk.
			// Workers have ~90-95Gi actually free. At 200% the 240Gi scheduling cap
			// was exhausted (no room for new 50Gi replicas). 300% gives ~150Gi
			// headroom per node for AI model volumes and future growth.
			"storageOverProvisioningPercentage": 300,
		},
		"persistence": map[string]any{
			"defaultClass":             true,
			"defaultClassReplicaCount": 3,
		},
		// Expose Longhorn metrics to VMAgent via a ServiceMonitor.
		"metrics": map[string]any{
			"serviceMonitor": map[string]any{
				"enabled": true,
			},
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

	// Gateway API HTTPRoute — routes longhorn.madhan.app → longhorn-frontend:80
	cdk8s.NewApiObject(chart, jsii.String("longhorn-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("longhorn"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"longhorn.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "longhorn-frontend", "port": 80, "weight": 1},
				},
			},
		},
	}))

	longhorn.NewLonghorn(chart, jsii.String("longhorn-release"), &longhorn.LonghornProps{
		ReleaseName: jsii.String("longhorn"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
