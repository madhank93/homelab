package automation

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewN8nChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Create namespace
	cdk8s.NewApiObject(chart, jsii.String("n8n-namespace"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Namespace"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String(namespace),
		},
	})

	// Create InfisicalSecret CRD
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
					"secretsPath": "/n8n",
				},
			},
		},
		"managedSecretReference": map[string]any{
			"secretName":      "n8n-db",
			"secretNamespace": namespace,
			"creationPolicy":  "Owner",
		},
	}

	cdk8s.NewApiObject(chart, jsii.String("n8n-infisical-secret"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets.infisical.com/v1alpha1"),
		Kind:       jsii.String("InfisicalSecret"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("n8n-secrets"),
			Namespace: jsii.String(namespace),
			// ServerSideApply=false: Infisical CRD schema omits projectSlug from
			// serviceToken.secretsScope, causing SSA schema validation to fail.
			// Use client-side apply for this resource only.
			Annotations: &map[string]*string{
				"argocd.argoproj.io/sync-options": jsii.String("ServerSideApply=false"),
			},
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), infisicalSpec))

	values := map[string]any{
		// Main node configuration (UI and API)
		"main": map[string]any{
			"count": 1,
			"resources": map[string]any{
				"requests": map[string]any{
					"cpu":    "100m",
					"memory": "256Mi",
				},
				"limits": map[string]any{
					"cpu":    "500m",
					"memory": "512Mi",
				},
			},
			// Persistence for main node
			"persistence": map[string]any{
				"enabled":    true,
				"size":       "10Gi",
				"mountPath":  "/home/node/.n8n",
				"accessMode": "ReadWriteOnce",
			},
			"extraEnvVars": map[string]any{
				"N8N_HOST": "n8n.madhan.app",
			},
			// No affinity field at all - let Kubernetes handle scheduling
		},

		// Service configuration
		"service": map[string]any{
			"enabled": true,
			"type":    "ClusterIP",
			"port":    5678,
		},

		// Ingress disabled — traffic routed via Gateway API HTTPRoute below
		"ingress": map[string]any{
			"enabled": false,
		},

		// Database configuration - PostgreSQL for production
		"db": map[string]any{
			"type": "postgresdb",
		},

		// Built-in PostgreSQL using Bitnami chart
		"postgresql": map[string]any{
			"enabled": true,
			"auth": map[string]any{
				"database":       "n8n",
				"username":       "n8n",
				"existingSecret": "n8n-db", // Secret created by InfisicalSecret
				"secretKeys": map[string]any{ // Key mapping
					"adminPasswordKey": "DB_PASSWORD",
					"userPasswordKey":  "DB_PASSWORD",
				},
			},
			"primary": map[string]any{
				"persistence": map[string]any{
					"enabled": true,
					"size":    "10Gi",
				},
			},
		},

		// Image configuration
		"image": map[string]any{
			"repository": "n8nio/n8n",
			"tag":        "1.78.0", // Pinned — never use 'latest' (violates versioning policy)
			"pullPolicy": "IfNotPresent",
		},

		// Security context
		"securityContext": map[string]any{
			"allowPrivilegeEscalation": false,
			"capabilities": map[string]any{
				"drop": []string{"ALL"},
			},
			"readOnlyRootFilesystem": false,
			"runAsNonRoot":           true,
			"runAsUser":              1000,
			"runAsGroup":             1000,
		},

		"podSecurityContext": map[string]any{
			"fsGroup":             1000,
			"fsGroupChangePolicy": "OnRootMismatch",
		},

		// Service account
		"serviceAccount": map[string]any{
			"create": true,
			"name":   "",
		},

		// Environment variables removed (moved to main.extraEnvVars)
	}

	cdk8s.NewHelm(chart, jsii.String("n8n-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("n8n"),
		Repo:        jsii.String("https://community-charts.github.io/helm-charts"),
		Version:     jsii.String("1.16.29"), // Bumped from 1.16.28 (released 2026-02-20)
		ReleaseName: jsii.String("n8n"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// Gateway API HTTPRoute — routes n8n.madhan.app → n8n-main:5678
	cdk8s.NewApiObject(chart, jsii.String("n8n-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("n8n"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"n8n.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"name": "n8n-main", "port": 5678},
				},
			},
		},
	}))

	return chart
}
