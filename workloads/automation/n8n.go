package automation

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewN8nChart deploys n8n workflow automation via the 8gears Helm chart (OCI).
//
// Secrets are sourced from OpenBao via the CSI Driver (Pattern B):
//   - ENCRYPTION_KEY is fetched from OpenBao and synced to k8s Secret "n8n-secrets".
//   - The PostgreSQL password is managed by CloudNativePG and stored in "n8n-pg-app",
//     from the CNPG Cluster created alongside.
func NewN8nChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, false)

	// DB password is owned by CloudNativePG ("n8n-pg-app"), not OpenBao.
	kit.SecretProviderClass(chart, "n8n-secrets", "n8n", []string{"ENCRYPTION_KEY"}, "n8n-secrets",
		kit.Sync{Object: "ENCRYPTION_KEY", Key: "N8N_ENCRYPTION_KEY"})

	// CloudNativePG Cluster — single-instance PostgreSQL for n8n.
	// CNPG auto-creates:
	//   - Secret  "n8n-pg-app"  → username/password for the "n8n" app user
	//   - Service "n8n-pg-rw"   → read-write endpoint (primary)
	// No manual password management required; CNPG owns the credential lifecycle.
	// imageName is the operand: a minor bump within one major is an in-place
	// rolling update, a major bump needs a new cluster or pg_upgrade.
	pg := config.ContainerImage("n8n-postgres")
	cdk8s.NewApiObject(chart, jsii.String("n8n-pg-cluster"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("postgresql.cnpg.io/v1"),
		Kind:       jsii.String("Cluster"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("n8n-pg"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"instances": 1,
		"imageName": pg.Image + ":" + pg.Tag,
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

	// main.config renders to a ConfigMap (non-sensitive); secrets come in through
	// main.extraEnv valueFrom: n8n-secrets (OpenBao) and n8n-pg-app (CNPG).
	values := map[string]any{
		"image": map[string]any{
			"tag": config.ContainerImage("n8n").Tag,
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
					"host": config.C.Host("n8n"),
					// n8n 2.x dropped in-memory binary data; keep it on the persistent volume.
					"default_binary_data_mode": "filesystem",
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
			// Major upgrades run DB migrations before n8n listens; allow 10 minutes
			// so the liveness probe does not kill a pod mid-migration.
			"startupProbe": map[string]any{
				"httpGet":          map[string]any{"path": "/healthz", "port": "http"},
				"periodSeconds":    10,
				"failureThreshold": 60,
			},
			// The limit covers the Code-node task runner, which n8n 2.x runs as a
			// child process inside this container.
			"resources": map[string]any{
				"requests": map[string]any{"cpu": "100m", "memory": "256Mi"},
				"limits":   map[string]any{"cpu": "500m", "memory": "1Gi"},
			},
			// The CSI mount is what makes the driver create the n8n-secrets Secret.
			"extraVolumes":      []map[string]any{kit.SecretsVolumeValues("n8n-secrets")},
			"extraVolumeMounts": []map[string]any{kit.SecretsMountValues()},
		},
		"ingress": map[string]any{
			"enabled": false,
		},
		// Disable embedded Bitnami PostgreSQL subchart — using CloudNativePG instead.
		"postgresql": map[string]any{
			"enabled": false,
		},
	}

	n8nChart := config.HelmChart("n8n")
	cdk8s.NewHelm(chart, jsii.String("n8n-release"), &cdk8s.HelmProps{
		// OCI charts take no Repo; the registry path is part of Chart.
		Chart:       jsii.String(n8nChart.Repo + "/n8n"),
		Version:     jsii.String(n8nChart.Version),
		ReleaseName: jsii.String("n8n"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// 8gears chart service: fullname = releaseName, port 80 → targetPort 5678.
	kit.Route(chart, "n8n", "n8n", "n8n", 80)

	return chart
}
