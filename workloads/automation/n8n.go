package automation

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

// NewN8nChart deploys n8n workflow automation via the 8gears Helm chart (OCI v2.0.1).
//
// Secrets are sourced from OpenBao via the CSI Driver (Pattern B):
//   - ENCRYPTION_KEY is fetched from OpenBao and synced to k8s Secret "n8n-secrets".
//   - The PostgreSQL password is managed by CloudNativePG and stored in "n8n-pg-app".
//
// A CNPG Cluster CR is also created in the same namespace to provision the
// n8n PostgreSQL database. The CSI volume mount on the n8n pod triggers secretObjects sync.
func NewN8nChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("n8n-namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
		},
	})

	// SecretProviderClass — fetches ENCRYPTION_KEY from OpenBao and syncs it into
	// k8s Secret "n8n-secrets" as N8N_ENCRYPTION_KEY.
	// DB password is managed by CloudNativePG (auto-generated in "n8n-pg-app" secret).
	// The CSI volume on the n8n pod triggers this sync.
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
			"objects": `- objectName: "ENCRYPTION_KEY"
  secretPath: "secret/data/n8n"
  secretKey: "ENCRYPTION_KEY"`,
		},
		"secretObjects": []map[string]any{
			{
				"secretName": "n8n-secrets",
				"type":       "Opaque",
				"data": []map[string]any{
					{"objectName": "ENCRYPTION_KEY", "key": "N8N_ENCRYPTION_KEY"},
				},
			},
		},
	}))

	// CloudNativePG Cluster — single-instance PostgreSQL for n8n.
	// CNPG auto-creates:
	//   - Secret  "n8n-pg-app"  → username/password for the "n8n" app user
	//   - Service "n8n-pg-rw"   → read-write endpoint (primary)
	// No manual password management required; CNPG owns the credential lifecycle.
	cdk8s.NewApiObject(chart, jsii.String("n8n-pg-cluster"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("postgresql.cnpg.io/v1"),
		Kind:       jsii.String("Cluster"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("n8n-pg"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"instances": 1,
		"storage": map[string]any{
			"size":         "10Gi",
			"storageClass": "longhorn",
		},
		"bootstrap": map[string]any{
			"initdb": map[string]any{
				"database": "n8n",
				"owner":    "n8n",
			},
		},
		"resources": map[string]any{
			"requests": map[string]any{"cpu": "100m", "memory": "256Mi"},
			"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
		},
	}))

	// n8n via 8gears Helm chart (OCI registry).
	//
	// Secrets strategy:
	//   N8N_ENCRYPTION_KEY   ← extraEnv.valueFrom → n8n-secrets    (CSI-synced from OpenBao)
	//   DB_POSTGRESDB_PASSWORD ← extraEnv.valueFrom → n8n-pg-app   (CNPG-generated)
	//
	// main.config → rendered as a ConfigMap; keys map to n8n env vars (non-sensitive).
	// main.extraEnv → individual env vars referencing existing k8s Secrets.
	// CSI volume mount required to trigger secretObjects sync for n8n-secrets.
	values := map[string]any{
		"image": map[string]any{
			"tag": "1.78.0", // Pinned — never use 'latest' (violates versioning policy)
		},
		"main": map[string]any{
			"config": map[string]any{
				"db": map[string]any{
					"type": "postgresdb",
					"postgresdb": map[string]any{
						// n8n-pg-rw: CNPG read-write service (primary)
						"host":     "n8n-pg-rw",
						"port":     5432,
						"user":     "n8n",
						"database": "n8n",
					},
				},
				"n8n": map[string]any{
					"host": "n8n.madhan.app",
				},
			},
			// Secrets injected from k8s Secrets via valueFrom — never stored in ConfigMap.
			"extraEnv": map[string]any{
				"N8N_ENCRYPTION_KEY": map[string]any{
					"valueFrom": map[string]any{
						"secretKeyRef": map[string]any{
							"name": "n8n-secrets",
							"key":  "N8N_ENCRYPTION_KEY",
						},
					},
				},
				"DB_POSTGRESDB_PASSWORD": map[string]any{
					"valueFrom": map[string]any{
						"secretKeyRef": map[string]any{
							// CNPG auto-creates this secret when the Cluster CR is reconciled.
							"name": "n8n-pg-app",
							"key":  "password",
						},
					},
				},
			},
			"persistence": map[string]any{
				"enabled":     true,
				"type":        "dynamic",
				"size":        "10Gi",
				"accessModes": []string{"ReadWriteMany"},
			},
			"replicaCount": 1,
			"resources": map[string]any{
				"requests": map[string]any{"cpu": "100m", "memory": "256Mi"},
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
			},
			// CSI volume triggers secretObjects sync → creates n8n-secrets Secret.
			// Required: secretObjects only sync when a pod mounts the CSI volume.
			"extraVolumes": []map[string]any{
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
			"extraVolumeMounts": []map[string]any{
				{
					"name":      "openbao-secrets",
					"mountPath": "/mnt/secrets",
					"readOnly":  true,
				},
			},
		},
		"ingress": map[string]any{
			"enabled": false,
		},
		// Disable embedded Bitnami PostgreSQL subchart — using CloudNativePG instead.
		"postgresql": map[string]any{
			"enabled": false,
		},
	}

	cdk8s.NewHelm(chart, jsii.String("n8n-release"), &cdk8s.HelmProps{
		// OCI registry — no Repo field for OCI charts
		Chart:       jsii.String("oci://8gears.container-registry.com/library/n8n"),
		Version:     jsii.String("2.0.1"),
		ReleaseName: jsii.String("n8n"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// Gateway API HTTPRoute — routes n8n.madhan.app → n8n:80
	// 8gears chart service: fullname = releaseName ("n8n"), port 80 → targetPort 5678
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
					{"group": "", "kind": "Service", "name": "n8n", "port": 80, "weight": 1},
				},
			},
		},
	}))

	return chart
}
