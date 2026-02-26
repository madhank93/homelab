package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewGrafanaChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Create namespace
	cdk8s.NewApiObject(chart, jsii.String("monitoring-namespace"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Namespace"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String(namespace),
		},
	})

	// Create InfisicalSecret CRD to sync secrets from Infisical
	infisicalSpec := map[string]any{
		"hostAPI":        "http://infisical-infisical-standalone-infisical.infisical.svc.cluster.local:8080",
		"resyncInterval": 60,
		"authentication": map[string]any{
			"serviceToken": map[string]any{
				"serviceTokenSecretReference": map[string]any{
					"secretName":      "infisical-service-token",
					"secretNamespace": "infisical",
				},
				"secretsScope": map[string]any{
					"projectSlug": "homelab-prod",
					"envSlug":     "prod",
					"secretsPath": "/grafana",
				},
			},
		},
		"managedSecretReference": map[string]any{
			"secretName":      "grafana-admin",
			"secretNamespace": namespace,
			"creationPolicy":  "Owner",
		},
	}

	cdk8s.NewApiObject(chart, jsii.String("grafana-infisical-secret"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets.infisical.com/v1alpha1"),
		Kind:       jsii.String("InfisicalSecret"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("grafana-secrets"),
			Namespace: jsii.String(namespace),
			// ServerSideApply=false: Infisical CRD schema omits projectSlug from
			// serviceToken.secretsScope, causing SSA schema validation to fail.
			// Use client-side apply for this resource only.
			Annotations: &map[string]*string{
				"argocd.argoproj.io/sync-options": jsii.String("ServerSideApply=false"),
			},
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), infisicalSpec))

	// Grafana Helm chart configuration
	values := map[string]any{
		// Datasources provisioned automatically — no manual UI setup required
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
						"url":    "http://victoria-logs-victoria-logs-single-server.victoria-logs.svc.cluster.local:9428/select/loki",
						"access": "proxy",
					},
				},
			},
		},
		"admin": map[string]any{
			"existingSecret": "grafana-admin",  // Secret created by InfisicalSecret
			"passwordKey":    "ADMIN_PASSWORD", // Key from Infisical
		},
		"resources": map[string]any{
			"limits": map[string]any{
				"cpu":    "500m",
				"memory": "512Mi",
			},
			"requests": map[string]any{
				"cpu":    "100m",
				"memory": "128Mi",
			},
		},
		"persistence": map[string]any{
			"enabled": true,
			"size":    "10Gi",
		},
		"service": map[string]any{
			"type": "ClusterIP",
			"port": 3000,
		},
		// Ingress disabled — traffic routed via Gateway API HTTPRoute below
		"ingress": map[string]any{
			"enabled": false,
		},
	}

	cdk8s.NewHelm(chart, jsii.String("grafana-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("grafana"),
		Repo:        jsii.String("https://grafana-community.github.io/helm-charts"),
		Version:     jsii.String("10.7.0"),
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
			{"name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"grafana.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"name": "grafana", "port": 3000},
				},
			},
		},
	}))

	return chart
}
