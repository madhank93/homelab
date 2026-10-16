package registry

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	harborimport "github.com/madhank93/homelab/workloads/imports/harbor"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewHarborChart deploys Harbor container registry via the Helm chart.
//
// The admin password comes from OpenBao (Pattern B). The Harbor chart cannot
// mount extra volumes, so a secret-sync pod mounts the CSI volume instead.
//
// The HTTPRoute must target the harbor:80 nginx proxy Service, not harbor-core:80.
//
// Known issue: RWO PVC deadlock on rolling update — force-delete pods and scale
// down the old ReplicaSet to zero to resolve.
func NewHarborChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, false)
	kit.SecretProviderClass(chart, "harbor-secrets", "harbor", []string{"HARBOR_ADMIN_PASSWORD"}, "harbor-admin",
		kit.Sync{Object: "HARBOR_ADMIN_PASSWORD", Key: "HARBOR_ADMIN_PASSWORD"})
	kit.SecretSync(chart, "harbor-secrets")

	values := map[string]any{
		"expose": map[string]any{
			"type": "clusterIP",
			"tls":  map[string]any{"enabled": false},
		},
		"externalURL": "https://" + config.C.Host("harbor"),
		"persistence": map[string]any{
			"enabled": true,
			// ReadWriteMany via Longhorn NFS — eliminates RWO multi-attach deadlocks
			// on rolling updates. Longhorn automatically provisions an NFS server pod
			// for each RWX volume.
			"persistentVolumeClaim": map[string]any{
				"registry": map[string]any{
					"size":       "50Gi",
					"accessMode": "ReadWriteMany",
				},
				"jobservice": map[string]any{
					"jobLog": map[string]any{
						"accessMode": "ReadWriteMany",
					},
				},
				"database": map[string]any{"size": "10Gi"},
			},
		},
		"harborAdminPassword": "",
		"existingSecret":      "harbor-admin", // Secret synced by SecretProviderClass + secret-sync pod
		"core": map[string]any{
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "1000m", "memory": "1Gi"},
				"requests": map[string]any{"cpu": "100m", "memory": "256Mi"},
			},
		},
		"jobservice": map[string]any{
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
		},
		"registry": map[string]any{
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "1000m", "memory": "1Gi"},
				"requests": map[string]any{"cpu": "100m", "memory": "256Mi"},
			},
		},
	}

	harborimport.NewHarbor(chart, jsii.String("harbor-release"), &harborimport.HarborProps{
		ReleaseName: jsii.String("harbor"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	kit.Route(chart, "harbor", "harbor", "harbor", 80)

	return chart
}
