package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/grafana"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewGrafanaChart deploys Grafana into the given namespace with OpenBao-sourced secrets
// and pre-provisioned VictoriaMetrics and VictoriaLogs datasources.
//
// Secrets use Pattern B (secretObjects sync) via the CSI Driver:
//   - ADMIN_PASSWORD: file-mounted at /mnt/secrets/ADMIN_PASSWORD, read via GF_SECURITY_ADMIN_PASSWORD__FILE.
//   - OAUTH_CLIENT_SECRET: synced to k8s Secret "grafana-oauth-secret" for Authentik OIDC.
//
// Kyverno policy engine dashboards are imported via a ConfigMap with the
// grafana_dashboard label so Grafana's sidecar picks them up automatically.
func NewGrafanaChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, false)

	// ADMIN_PASSWORD is only file-mounted; OAUTH_CLIENT_SECRET is mirrored into
	// a Secret so Grafana can read it as a GF_ env var.
	kit.SecretProviderClass(chart, "grafana-secrets", "grafana",
		[]string{"ADMIN_PASSWORD", "OAUTH_CLIENT_SECRET"}, "grafana-oauth-secret",
		kit.Sync{Object: "OAUTH_CLIENT_SECRET", Key: "GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET"})

	auth := "https://" + config.C.Host("auth")
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
						"url":       "http://vmsingle-vm-stack.victoria-metrics.svc.cluster.local:8428",
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
						// VictoriaLogs does not implement the Loki index/volume API.
						// Disabling logsVolumeEnabled suppresses the error in Drilldown Logs.
						"jsonData": map[string]any{
							"logsVolumeEnabled": false,
						},
					},
				},
			},
		},
		// Authentik OIDC — GitHub login flows through Authentik.
		// Steps to activate:
		//   1. Authentik UI → Directory → Social Logins → Add GitHub source (slug: github)
		//   2. Authentik UI → Applications → Providers → Create OAuth2/OIDC provider
		//        Name: Grafana | Slug: grafana | Scopes: openid, email, profile
		//        Redirect URI: https://grafana.<domain>/login/generic_oauth
		//   3. Authentik UI → Applications → Create Application → bind to Grafana provider
		//   4. Copy the Client ID from the provider and replace the placeholder below.
		//   5. Copy the Client Secret and store it in OpenBao:
		//        bao kv patch secret/grafana OAUTH_CLIENT_SECRET=<secret>
		//   6. In Authentik, create a group "grafana-admins" and add yourself to it for Admin role.
		"grafana.ini": map[string]any{
			// root_url must be set so Grafana constructs OAuth redirect_uri with https://.
			// Without it, Grafana uses the incoming request scheme (http — TLS terminates at
			// Bifrost/Traefik) and sends http://grafana.<domain>/login/generic_oauth,
			// which mismatches the https:// redirect_uri registered in Authentik.
			"server": map[string]any{
				"root_url": "https://" + config.C.Host("grafana"),
			},
			"auth.generic_oauth": map[string]any{
				"enabled": true,
				"name":    "GitHub via Authentik",
				// Client ID set in core/cloud/authentik.go — matches GrafanaOIDCClientID export.
				"client_id":     "grafana-homelab",
				"scopes":        "openid email profile",
				"auth_url":      auth + "/application/o/authorize/",
				"token_url":     auth + "/application/o/token/",
				"api_url":       auth + "/application/o/userinfo/",
				"use_pkce":      true,
				"allow_sign_up": true,
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
		"extraVolumes":      []map[string]any{kit.SecretsVolumeValues("grafana-secrets")},
		"extraVolumeMounts": []map[string]any{kit.SecretsMountValues()},
	}

	grafana.NewGrafana(chart, jsii.String("grafana-release"), &grafana.GrafanaProps{
		ReleaseName: jsii.String("grafana"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	kit.Route(chart, "grafana", "grafana", "grafana", 3000)

	return chart
}
