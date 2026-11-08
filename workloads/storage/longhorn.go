// Package storage holds the Longhorn distributed block storage chart.
package storage

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/longhorn"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewLonghornChart deploys Longhorn distributed block storage, the default
// StorageClass, on the worker disks sized in config/config.yml. The namespace is
// privileged because the CSI driver and instance managers need host devices.
func NewLonghornChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, true)

	values := map[string]any{
		"defaultSettings": map[string]any{
			"defaultReplicaCount":           3,
			"defaultDataPath":               "/var/lib/longhorn/", // Talos persistent path
			"createDefaultDiskLabeledNodes": false,                // default disk on every Longhorn node, not only labelled ones
			// Volumes only consume what they write, so claims may exceed free space,
			// leaving room for 50Gi AI model replicas.
			"storageOverProvisioningPercentage": 300,
			// Moves volume engines onto the chart's engine image after a Longhorn
			// upgrade without manual per-volume upgrades; one volume per node at a time.
			"concurrentAutomaticEngineUpgradePerNodeLimit": 1,
			// Lets the components Longhorn deploys itself (CSI plugin, engine image,
			// instance and share managers) run on the GPU worker. Changing it restarts
			// them, and Longhorn defers parts of it while volumes are attached.
			"taintToleration": "dedicated=ai:NoSchedule",
			// Force-deletes pods on a node that went down uncleanly, so their RWO
			// volumes detach and the pods reschedule instead of hanging in Terminating.
			"nodeDownPodDeletionPolicy": "delete-both-statefulset-and-deployment-pod",
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
		// Argo CD runs this Helm pre-upgrade hook on every sync, including the first install.
		"preUpgradeChecker": map[string]any{
			"jobEnabled": false,
		},
		// longhorn-manager must not tolerate the control-plane taint: a manager pod
		// registers its node as a Longhorn node, and every registered node needs an
		// engine image. Engine-image DaemonSets get no tolerations, so control planes
		// would leave EngineImages stuck "deploying" and block volume attachment.
		// The GPU worker's dedicated=ai taint is the exception: taintToleration
		// above gives engine images the matching toleration.
		"longhornManager": map[string]any{
			"tolerations": []map[string]any{
				{"key": "dedicated", "operator": "Equal", "value": "ai", "effect": "NoSchedule"},
			},
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
		// Helm binary's built-in default rather than the cluster's version.
		HelmFlags: &[]*string{jsii.String("--kube-version=" + config.V.Kubernetes.Version)},
	})

	return chart
}
