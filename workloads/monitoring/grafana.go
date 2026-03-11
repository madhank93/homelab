package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/grafana"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

func NewGrafanaChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("monitoring-namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
		},
	})

	// SecretProviderClass — Pattern B (secretObjects sync).
	// Mounts secrets from OpenBao and syncs selected keys to a k8s Secret.
	//   ADMIN_PASSWORD      → file at /mnt/secrets/ADMIN_PASSWORD (read via __FILE env var)
	//   OAUTH_CLIENT_SECRET → synced to k8s Secret grafana-oauth-secret (read as GF_ env var)
	cdk8s.NewApiObject(chart, jsii.String("grafana-spc"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets-store.csi.x-k8s.io/v1"),
		Kind:       jsii.String("SecretProviderClass"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("grafana-secrets"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"provider": "openbao",
		"parameters": map[string]any{
			"vaultAddress": "http://openbao.openbao.svc.cluster.local:8200",
			"roleName":     "grafana",
			"objects": `- objectName: "ADMIN_PASSWORD"
  secretPath: "secret/data/grafana"
  secretKey: "ADMIN_PASSWORD"
- objectName: "OAUTH_CLIENT_SECRET"
  secretPath: "secret/data/grafana"
  secretKey: "OAUTH_CLIENT_SECRET"`,
		},
		// Sync OAUTH_CLIENT_SECRET to a k8s Secret so Grafana can read it as a GF_ env var.
		// The CSI volume mount (below) must be present to trigger the sync.
		"secretObjects": []map[string]any{
			{
				"secretName": "grafana-oauth-secret",
				"type":       "Opaque",
				"data": []map[string]any{
					{"objectName": "OAUTH_CLIENT_SECRET", "key": "GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET"},
				},
			},
		},
	}))

	values := map[string]any{
		"podAnnotations": map[string]any{
			"reloader.stakater.com/auto": "true",
		},
		"datasources": map[string]any{
			"datasources.yaml": map[string]any{
				"apiVersion": 1,
				"datasources": []map[string]any{
					{
						"name":      "VictoriaMetrics",
						"type":      "prometheus",
						"url":       "http://victoria-metrics-victoria-metrics-cluster-vmselect.victoria-metrics.svc.cluster.local:8481/select/0/prometheus",
						"access":    "proxy",
						"isDefault": true,
						"jsonData": map[string]any{
							"timeInterval": "30s",
						},
					},
					{
						"name":   "VictoriaLogs",
						"type":   "loki",
						"url":    "http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/select",
						"access": "proxy",
					},
				},
			},
		},
		// Authentik OIDC — GitHub login flows through Authentik.
		// Steps to activate:
		//   1. Authentik UI → Directory → Social Logins → Add GitHub source (slug: github)
		//   2. Authentik UI → Applications → Providers → Create OAuth2/OIDC provider
		//        Name: Grafana | Slug: grafana | Scopes: openid, email, profile
		//        Redirect URI: https://grafana.madhan.app/login/generic_oauth
		//   3. Authentik UI → Applications → Create Application → bind to Grafana provider
		//   4. Copy the Client ID from the provider and replace the placeholder below.
		//   5. Copy the Client Secret and store it in OpenBao:
		//        bao kv patch secret/grafana OAUTH_CLIENT_SECRET=<secret>
		//   6. In Authentik, create a group "grafana-admins" and add yourself to it for Admin role.
		"grafana.ini": map[string]any{
			"auth.generic_oauth": map[string]any{
				"enabled": true,
				"name":    "GitHub via Authentik",
				// Client ID set in core/cloud/authentik.go — matches GrafanaOIDCClientID export.
				"client_id":            "grafana-homelab",
				"scopes":               "openid email profile",
				"auth_url":             "https://auth.madhan.app/application/o/grafana/authorize/",
				"token_url":            "https://auth.madhan.app/application/o/grafana/token/",
				"api_url":              "https://auth.madhan.app/application/o/userinfo/",
				"use_pkce":             true,
				"allow_sign_up":        true,
				// Members of the grafana-admins Authentik group get Admin role; everyone else Viewer.
				"role_attribute_path": "contains(groups[*], 'grafana-admins') && 'Admin' || 'Viewer'",
			},
			// Disable Grafana's built-in login form for non-admin users.
			// Admin can still reach /login for username+password access.
			"auth": map[string]any{
				"disable_login_form": false,
			},
		},
		// Client secret injected from grafana-oauth-secret k8s Secret (synced by CSI driver above).
		// GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET overrides auth.generic_oauth.client_secret in ini.
		"envFromSecret": "grafana-oauth-secret",
		// Admin credentials: user set directly in env; password read from CSI-mounted file.
		// The Grafana chart's auto-generated admin Secret is NOT used to avoid synth-time
		// random value churn. GF_SECURITY_ADMIN_PASSWORD__FILE takes precedence over the Secret.
		"env": map[string]any{
			"GF_SECURITY_ADMIN_USER":           "admin",
			"GF_SECURITY_ADMIN_PASSWORD__FILE": "/mnt/secrets/ADMIN_PASSWORD",
		},
		"resources": map[string]any{
			"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
			"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
		},
		"persistence": map[string]any{
			"enabled":     true,
			"size":        "10Gi",
			"accessModes": []string{"ReadWriteMany"},
		},
		"service": map[string]any{
			"type": "ClusterIP",
			"port": 3000,
		},
		"ingress": map[string]any{"enabled": false},
		// Sidecar watches for ConfigMaps with label grafana_dashboard="1" across
		// all namespaces and auto-provisions them as dashboards.
		"sidecar": map[string]any{
			"dashboards": map[string]any{
				"enabled":          true,
				"searchNamespace":  "ALL",
				"label":            "grafana_dashboard",
				"labelValue":       "1",
				"folderAnnotation": "grafana_folder",
				"provider": map[string]any{
					"foldersFromFilesStructure": true,
				},
			},
		},
		"extraVolumes": []map[string]any{
			{
				"name": "openbao-secrets",
				"csi": map[string]any{
					"driver":   "secrets-store.csi.k8s.io",
					"readOnly": true,
					"volumeAttributes": map[string]any{
						"secretProviderClass": "grafana-secrets",
					},
				},
			},
		},
		"extraVolumeMounts": []map[string]any{
			{
				"name":      "openbao-secrets",
				"mountPath": "/mnt/secrets",
				"readOnly":  true,
			},
		},
	}

	grafana.NewGrafana(chart, jsii.String("grafana-release"), &grafana.GrafanaProps{
		ReleaseName: jsii.String("grafana"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// Gateway API HTTPRoute — routes grafana.madhan.app → grafana:3000
	cdk8s.NewApiObject(chart, jsii.String("grafana-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("grafana"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"grafana.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "grafana", "port": 3000, "weight": 1},
				},
			},
		},
	}))

	return chart
}
