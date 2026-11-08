// Package management deploys cluster management UIs: Headlamp and Homepage.
package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/headlamp"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewHeadlampChart deploys Headlamp, a Kubernetes web UI, via its Helm chart.
// Users log in through Authentik OIDC; a cluster-admin ServiceAccount is the
// token-login fallback.
func NewHeadlampChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// The client secret is synced from OpenBao; the rest is plain env.
	kit.SecretProviderClass(chart, "headlamp-secrets", "headlamp", []string{"OIDC_CLIENT_SECRET"}, "headlamp-oidc",
		kit.Sync{Object: "OIDC_CLIENT_SECRET", Key: "OIDC_CLIENT_SECRET"})

	headlamp.NewHeadlamp(chart, jsii.String("headlamp-release"), &headlamp.HeadlampProps{
		ReleaseName: jsii.String("headlamp"),
		Namespace:   jsii.String(namespace),
		Values: &headlamp.HeadlampValues{
			AdditionalValues: &map[string]any{
				"resources": map[string]any{
					"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
					"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
				},
				// Exposed through the Gateway instead.
				"ingress": map[string]any{"enabled": false},
				// Authentik login. Headlamp passes the ID token to the API server, which
				// trusts the "kubernetes" client (apiServer extraArgs in core/platform/talos.go).
				"config": map[string]any{
					"oidc": map[string]any{
						"secret":         map[string]any{"create": false},
						"externalSecret": map[string]any{"enabled": true, "name": "headlamp-oidc", "hasScopes": true},
					},
				},
				"env": []map[string]any{
					{"name": "OIDC_CLIENT_ID", "value": "kubernetes"},
					{"name": "OIDC_ISSUER_URL", "value": "https://" + config.C.Host("auth") + "/application/o/kubernetes/"},
					{"name": "OIDC_SCOPES", "value": "openid,profile,email,groups,offline_access"},
				},
				"volumes":      []map[string]any{kit.SecretsVolumeValues("headlamp-secrets")},
				"volumeMounts": []map[string]any{kit.SecretsMountValues()},
			},
		},
	})

	// Login identity for the UI. No static token Secret: mint a short-lived one
	// with `just headlamp-token` (kubectl create token).
	k8s.NewKubeServiceAccount(chart, jsii.String("headlamp-admin-sa"), &k8s.KubeServiceAccountProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("headlamp-admin"),
			Namespace: jsii.String(namespace),
		},
	})

	k8s.NewKubeClusterRoleBinding(chart, jsii.String("headlamp-admin-crb"), &k8s.KubeClusterRoleBindingProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String("headlamp-admin"),
		},
		RoleRef: &k8s.RoleRef{
			ApiGroup: jsii.String("rbac.authorization.k8s.io"),
			Kind:     jsii.String("ClusterRole"),
			Name:     jsii.String("cluster-admin"),
		},
		Subjects: &[]*k8s.Subject{
			{
				Kind:      jsii.String("ServiceAccount"),
				Name:      jsii.String("headlamp-admin"),
				Namespace: jsii.String(namespace),
			},
		},
	})

	// API server identities from Authentik carry the "oidc:" prefix (core/platform/talos.go).
	k8s.NewKubeClusterRoleBinding(chart, jsii.String("oidc-admins-crb"), &k8s.KubeClusterRoleBindingProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("oidc-grafana-admins")},
		RoleRef: &k8s.RoleRef{
			ApiGroup: jsii.String("rbac.authorization.k8s.io"),
			Kind:     jsii.String("ClusterRole"),
			Name:     jsii.String("cluster-admin"),
		},
		Subjects: &[]*k8s.Subject{
			{Kind: jsii.String("Group"), Name: jsii.String("oidc:grafana-admins"), ApiGroup: jsii.String("rbac.authorization.k8s.io")},
		},
	})

	kit.Route(chart, "headlamp", "headlamp", "headlamp", 80)

	return chart
}
