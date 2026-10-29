package monitoring

import (
	"fmt"
	"strings"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/grafana"
	"github.com/madhank93/homelab/workloads/kit"
)

// publicOrgID is the "Public" org every SSO user lands in as Viewer. Grafana
// cannot provision orgs, and refuses to start while provisioning references a
// missing one: `just grafana-init` creates it and asserts this ID.
const publicOrgID = 2

// victoriaMetrics is the metrics datasource, provisioned once per org.
func victoriaMetrics(orgID int) map[string]any {
	return map[string]any{
		"name":      "VictoriaMetrics",
		"orgId":     orgID,
		"type":      "prometheus",
		"url":       "http://vmsingle-vm-stack.victoria-metrics.svc.cluster.local:8428",
		"access":    "proxy",
		"isDefault": true,
		"jsonData": map[string]any{
			"timeInterval": "30s",
		},
	}
}

// NewGrafanaChart deploys Grafana into the given namespace with OpenBao-sourced secrets.
// The main org (admins only) gets VictoriaMetrics and VictoriaLogs; the Public org
// gets VictoriaMetrics and a few dashboards, never logs.
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

	// The chart renders "<registry>/<repository>:<tag>", so the pinned image
	// is split at its registry host.
	img := config.ContainerImage("grafana")
	registry, repository, _ := strings.Cut(img.Image, "/")

	auth := "https://" + config.C.Host("auth")
	values := map[string]any{
		"image": map[string]any{
			"registry":   registry,
			"repository": repository,
			"tag":        img.Tag,
		},
		"podAnnotations": map[string]any{
			"reloader.stakater.com/auto": "true",
		},
		"datasources": map[string]any{
			"datasources.yaml": map[string]any{
				"apiVersion": 1,
				"datasources": []map[string]any{
					victoriaMetrics(1),
					victoriaMetrics(publicOrgID),
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
				"scopes":        "openid email profile groups",
				"auth_url":      auth + "/application/o/authorize/",
				"token_url":     auth + "/application/o/token/",
				"api_url":       auth + "/application/o/userinfo/",
				"use_pkce":      true,
				"allow_sign_up": true,
				// grafana-admins → Admin of the main org; everyone → Viewer of Public.
				// The direct role only raises roles in mapped orgs, so it never grants org 1.
				"org_attribute_path":    "groups",
				"org_mapping":           fmt.Sprintf("grafana-admins:1:Admin *:%d:Viewer", publicOrgID),
				"role_attribute_path":   "contains(groups[*], 'grafana-admins') && 'Admin' || 'Viewer'",
				"role_attribute_strict": true,
			},
			// SSO only. Break-glass: `grafana-cli admin reset-admin-password`, then the
			// HTTP API with basic auth (still enabled) as the local admin.
			"auth": map[string]any{
				"disable_login_form": true,
			},
		},
		// Client secret injected from grafana-oauth-secret k8s Secret (synced by CSI driver above).
		// GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET overrides auth.generic_oauth.client_secret in ini.
		"envFromSecret": "grafana-oauth-secret",
		// Admin password read from the CSI-mounted file via Grafana's $__file provider;
		// setting it here also stops the chart rendering a random admin Secret on every
		// synth. The distroless image has no /run.sh, so GF_*__FILE vars are unsupported.
		"env": map[string]any{
			"GF_SECURITY_ADMIN_USER":     "admin",
			"GF_SECURITY_ADMIN_PASSWORD": "$__file{/mnt/secrets/ADMIN_PASSWORD}",
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
		// Helm test hooks have no meaning under Argo CD.
		"testFramework": map[string]any{"enabled": false},
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
		// Public org dashboards, downloaded from grafana.com by the chart's init
		// container. Sidecar ConfigMaps all land in the main org.
		"dashboardProviders": map[string]any{
			"dashboardproviders.yaml": map[string]any{
				"apiVersion": 1,
				"providers": []map[string]any{{
					"name":    "public",
					"orgId":   publicOrgID,
					"type":    "file",
					"options": map[string]any{"path": "/var/lib/grafana/dashboards/public"},
				}},
			},
		},
		"defaultCurlOptions": "-sf", // chart default adds -k (skips TLS verification)
		"dashboards": map[string]any{
			"public": map[string]any{
				"kubernetes-global": map[string]any{"gnetId": 15757, "revision": 43},
				"nvidia-dcgm": map[string]any{
					"gnetId":     12239,
					"revision":   2,
					"datasource": []map[string]any{{"name": "DS_PROMETHEUS", "value": "VictoriaMetrics"}},
				},
			},
		},
		"extraVolumes":      []map[string]any{kit.SecretsVolumeValues("grafana-secrets")},
		"extraVolumeMounts": []map[string]any{kit.SecretsMountValues()},
	}

	grafana.NewGrafana(chart, jsii.String("grafana-release"), &grafana.GrafanaProps{
		ReleaseName: jsii.String("grafana"),
		Namespace:   jsii.String(namespace),
		Values:      &grafana.GrafanaValues{AdditionalValues: &values},
	})

	kit.Route(chart, "grafana", "grafana", "grafana", 3000)

	return chart
}
