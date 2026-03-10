package automation

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

func NewN8nChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("n8n-namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
		},
	})

	// SecretProviderClass — Pattern B (secretObjects sync).
	// Fetches DB_PASSWORD from OpenBao and syncs it into k8s Secret "n8n-db".
	// N8n's bundled PostgreSQL subchart references existingSecret: "n8n-db".
	// The CSI volume on the n8n main pod triggers this sync.
	cdk8s.NewApiObject(chart, jsii.String("n8n-spc"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets-store.csi.x-k8s.io/v1"),
		Kind:       jsii.String("SecretProviderClass"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("n8n-secrets"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"provider": "openbao",
		"parameters": map[string]any{
			"vaultAddress": "http://openbao.openbao.svc.cluster.local:8200",
			"roleName":     "n8n",
			"objects": `- objectName: "DB_PASSWORD"
  secretPath: "secret/data/n8n"
  secretKey: "DB_PASSWORD"`,
		},
		"secretObjects": []map[string]any{
			{
				"secretName": "n8n-db",
				"type":       "Opaque",
				"data": []map[string]any{
					{
						"objectName": "DB_PASSWORD",
						"key":        "DB_PASSWORD",
					},
				},
			},
		},
	}))

	values := map[string]any{
		// Main node configuration (UI and API)
		"main": map[string]any{
			"count": 1,
			"podAnnotations": map[string]any{
				"reloader.stakater.com/auto": "true",
			},
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
			// CSI volume triggers secretObjects sync → creates n8n-db Secret.
			// Required: secretObjects only sync when a pod mounts the CSI volume.
			// n8n chart uses "volumes"/"volumeMounts" (not "extraVolumes"/"extraVolumeMounts")
			"volumes": []map[string]any{
				{
					"name": "openbao-secrets",
					"csi": map[string]any{
						"driver":   "secrets-store.csi.k8s.io",
						"readOnly": true,
						"volumeAttributes": map[string]any{
							"secretProviderClass": "n8n-secrets",
						},
					},
				},
			},
			"volumeMounts": []map[string]any{
				{
					"name":      "openbao-secrets",
					"mountPath": "/mnt/secrets",
					"readOnly":  true,
				},
			},
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
				"existingSecret": "n8n-db", // Secret synced by SecretProviderClass
				"secretKeys": map[string]any{
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
	}

	// NOTE: n8n typed import has many JSII-enforced required fields in N8nValues
	// (e.g. Affinity, Api, BinaryData...) that must all be populated to use
	// the typed construct. Using cdk8s.NewHelm to pass values as a plain map.
	cdk8s.NewHelm(chart, jsii.String("n8n-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("n8n"),
		Repo:        jsii.String("https://community-charts.github.io/helm-charts"),
		Version:     jsii.String("1.16.29"),
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
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"n8n.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "n8n-main", "port": 5678, "weight": 1},
				},
			},
		},
	}))

	return chart
}
