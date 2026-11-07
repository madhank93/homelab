package registry

import (
	"encoding/json"

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
func NewHarborChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, false)
	kit.SecretProviderClass(chart, "harbor-secrets", "harbor", []string{"HARBOR_ADMIN_PASSWORD", "OIDC_CLIENT_SECRET"}, "harbor-admin",
		kit.Sync{Object: "HARBOR_ADMIN_PASSWORD", Key: "HARBOR_ADMIN_PASSWORD"},
		kit.Sync{Object: "OIDC_CLIENT_SECRET", Key: "OIDC_CLIENT_SECRET"})

	// Harbor keeps auth settings in its database; CONFIG_OVERWRITE_JSON makes
	// core re-apply them on every start. The $(...) reference is expanded by
	// the kubelet, so the client secret stays out of the rendered manifest.
	// The local admin can still sign in at /account/sign-in.
	oidc, err := json.Marshal(map[string]any{ //nolint:gosec // client secret is a $(VAR) reference expanded by the kubelet
		"auth_mode":          "oidc_auth",
		"oidc_name":          "Authentik",
		"oidc_endpoint":      "https://" + config.C.Host("auth") + "/application/o/harbor/",
		"oidc_client_id":     "harbor-homelab",
		"oidc_client_secret": "$(OIDC_CLIENT_SECRET)",
		"oidc_scope":         "openid,profile,email,groups,offline_access",
		"oidc_verify_cert":   true,
		"oidc_auto_onboard":  true,
		"oidc_user_claim":    "preferred_username",
		"oidc_groups_claim":  "groups",
		"oidc_admin_group":   "grafana-admins",
	})
	if err != nil {
		panic(err)
	}
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
		"harborAdminPassword":         "",
		"existingSecretAdminPassword": "harbor-admin", // synced from OpenBao by the secret-sync pod
		"core": map[string]any{
			"extraEnvVars": []map[string]any{
				{"name": "OIDC_CLIENT_SECRET", "valueFrom": map[string]any{
					"secretKeyRef": map[string]any{"name": "harbor-admin", "key": "OIDC_CLIENT_SECRET", "optional": true},
				}},
				{"name": "CONFIG_OVERWRITE_JSON", "value": string(oidc)},
			},
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
