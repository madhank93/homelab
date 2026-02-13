package storage

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewLonghornChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Create namespace first with PSA labels for privileged workloads
	cdk8s.NewApiObject(chart, jsii.String("longhorn-namespace"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Namespace"),
		Metadata: &cdk8s.ApiObjectMetadata{
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
			"createDefaultDiskLabeledNodes": true,
		},
		"persistence": map[string]any{
			"defaultClass":             true,
			"defaultClassReplicaCount": 3,
		},
		// Disable pre-upgrade hook for initial installation (ArgoCD/GitOps best practice)
		"preUpgradeChecker": map[string]any{
			"jobEnabled": false,
		},
		// Talos-specific: Avoid control-plane nodes using tolerations
		"longhornManager": map[string]any{
			"tolerations": []map[string]any{
				{
					"key":      "node-role.kubernetes.io/control-plane",
					"operator": "Exists",
					"effect":   "NoSchedule",
				},
			},
		},
		"longhornDriver": map[string]any{
			"tolerations": []map[string]any{
				{
					"key":      "node-role.kubernetes.io/control-plane",
					"operator": "Exists",
					"effect":   "NoSchedule",
				},
			},
		},
		"longhornUI": map[string]any{
			"tolerations": []map[string]any{
				{
					"key":      "node-role.kubernetes.io/control-plane",
					"operator": "Exists",
					"effect":   "NoSchedule",
				},
			},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("longhorn-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("longhorn"),
		Repo:        jsii.String("https://charts.longhorn.io"),
		Version:     jsii.String("1.8.0"),
		ReleaseName: jsii.String("longhorn"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
