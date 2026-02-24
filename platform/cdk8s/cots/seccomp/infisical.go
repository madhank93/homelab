package seccomp

import (
	"os"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewInfisicalChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Create namespace first
	cdk8s.NewApiObject(chart, jsii.String("infisical-namespace"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Namespace"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String(namespace),
		},
	})

	// Get Infisical DB password from environment (only secret in GitHub)
	infisicalDbPassword := os.Getenv("INFISICAL_DB_PASSWORD")
	if infisicalDbPassword == "" {
		panic("INFISICAL_DB_PASSWORD environment variable is required")
	}

	// Get Infisical Encryption Key (must be 16 bytes / 32 hex chars)
	infisicalEncryptionKey := os.Getenv("INFISICAL_ENCRYPTION_KEY")
	if infisicalEncryptionKey == "" {
		panic("INFISICAL_ENCRYPTION_KEY environment variable is required")
	}
	// Basic validation for key length (16 bytes hex = 32 chars)
	if len(infisicalEncryptionKey) != 32 {
		panic("INFISICAL_ENCRYPTION_KEY must be a 16-byte hex string (32 characters)")
	}

	// Get Infisical Auth Secret
	infisicalAuthSecret := os.Getenv("INFISICAL_AUTH_SECRET")
	if infisicalAuthSecret == "" {
		panic("INFISICAL_AUTH_SECRET environment variable is required")
	}

	// Get Redis password for internal Redis auth
	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		panic("REDIS_PASSWORD environment variable is required")
	}

	// Create Secret for PostgreSQL password (will be sealed by CI)
	postgresSecret := cdk8s.NewApiObject(chart, jsii.String("infisical-postgresql-secret"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Secret"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("infisical-secrets"),
			Namespace: jsii.String(namespace),
		},
	})
	// Create full connection URI from password (to override Chart default which defaults to 'root')
	// URI Format: postgresql://infisical:<password>@postgresql:5432/infisical
	// Note: failing to provide this explicitly often causes the chart to use 'root'
	dbConnectionUri := "postgresql://infisical:" + infisicalDbPassword + "@postgresql:5432/infisical"

	postgresSecret.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/stringData"), map[string]string{
		"DB_PASSWORD":       infisicalDbPassword,
		"AUTH_SECRET":       infisicalAuthSecret,
		"ENCRYPTION_KEY":    infisicalEncryptionKey,
		"DB_CONNECTION_URI": dbConnectionUri, // Pre-calculated URI
	}))

	// Infisical standalone chart values
	values := map[string]any{
		"infisical": map[string]any{
			"kubeSecretRef": "infisical-secrets",
			"replicaCount":  1,
			"resources": map[string]any{
				"requests": map[string]any{
					"cpu":    "200m",
					"memory": "512Mi",
				},
				"limits": map[string]any{
					"memory": "1024Mi",
				},
			},
		},
		"postgresql": map[string]any{
			"enabled": false, // Disable embedded Postgres to handle it separately
			"useExistingPostgresSecret": map[string]any{
				"enabled": true,
				"existingConnectionStringSecret": map[string]any{
					"name": "infisical-secrets",
					"key":  "DB_CONNECTION_URI",
				},
			},
		},
		"redis": map[string]any{
			"enabled": true,
			"auth": map[string]any{
				"enabled":  true,
				"password": redisPassword,
			},
		},
		"ingress": map[string]any{
			"enabled":  false,
			"hostname": "infisical.madhan.app",
		},
		// Disable the bundled nginx subchart entirely.
		// Even with ingress.enabled=false the subchart still renders a Deployment
		// and a ValidatingWebhookConfiguration that can block all pod creation.
		"ingress-nginx": map[string]any{
			"enabled": false,
		},
	}

	// Deploy Infisical Application
	cdk8s.NewHelm(chart, jsii.String("infisical-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("infisical-standalone"),
		Repo:        jsii.String("https://dl.cloudsmith.io/public/infisical/helm-charts/helm/charts"),
		Version:     jsii.String("1.7.2"),
		ReleaseName: jsii.String("infisical"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// Deploy PostgreSQL using official image (docker.io/library/postgres:17).
	// Bitnami images were removed from Docker Hub and are no longer publicly pullable.
	// The official postgres image uses POSTGRES_USER/PASSWORD/DB env vars.
	cdk8s.NewApiObject(chart, jsii.String("postgresql-service"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Service"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("postgresql"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"selector": map[string]any{"app": "postgresql"},
		"ports": []map[string]any{
			{"port": 5432, "targetPort": 5432, "name": "postgres"},
		},
	}))

	cdk8s.NewApiObject(chart, jsii.String("postgresql-statefulset"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("apps/v1"),
		Kind:       jsii.String("StatefulSet"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("postgresql"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"serviceName": "postgresql",
		"replicas":    1,
		"selector": map[string]any{
			"matchLabels": map[string]any{"app": "postgresql"},
		},
		"template": map[string]any{
			"metadata": map[string]any{
				"labels": map[string]any{"app": "postgresql"},
			},
			"spec": map[string]any{
				"containers": []map[string]any{
					{
						"name":  "postgresql",
						"image": "docker.io/library/postgres:17",
						"ports": []map[string]any{
							{"containerPort": 5432, "name": "postgres"},
						},
						"env": []map[string]any{
							{"name": "POSTGRES_USER", "value": "infisical"},
							{"name": "POSTGRES_DB", "value": "infisical"},
							{
								"name": "POSTGRES_PASSWORD",
								"valueFrom": map[string]any{
									"secretKeyRef": map[string]any{
										"name": "infisical-secrets",
										"key":  "DB_PASSWORD",
									},
								},
							},
							{"name": "PGDATA", "value": "/var/lib/postgresql/data/pgdata"},
						},
						"resources": map[string]any{
							"requests": map[string]any{"cpu": "100m", "memory": "256Mi"},
							"limits":   map[string]any{"memory": "512Mi"},
						},
						"volumeMounts": []map[string]any{
							{"name": "data", "mountPath": "/var/lib/postgresql/data"},
						},
					},
				},
			},
		},
		"volumeClaimTemplates": []map[string]any{
			{
				"metadata": map[string]any{"name": "data"},
				"spec": map[string]any{
					"accessModes":      []string{"ReadWriteOnce"},
					"storageClassName": "longhorn",
					"resources": map[string]any{
						"requests": map[string]any{"storage": "10Gi"},
					},
				},
			},
		},
	}))

	// Gateway API HTTPRoute for Infisical
	cdk8s.NewApiObject(chart, jsii.String("infisical-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("infisical"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{
				"name":      "homelab-gateway",
				"namespace": "kube-system",
			},
		},
		"hostnames": []string{"infisical.madhan.app", "infisical.local"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{
						"name": "infisical-infisical-standalone-infisical",
						"port": 8080,
					},
				},
			},
		},
	}))

	// Infisical Operator (syncs secrets from Infisical to K8s)
	cdk8s.NewHelm(chart, jsii.String("infisical-operator-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("secrets-operator"),
		Repo:        jsii.String("https://dl.cloudsmith.io/public/infisical/helm-charts/helm/charts/"),
		Version:     jsii.String("0.10.23"),
		ReleaseName: jsii.String("infisical-operator"),

		Values: &map[string]any{
			"controllerManager": map[string]any{
				"manager": map[string]any{
					"resources": map[string]any{
						"limits": map[string]any{
							"cpu":    "500m",
							"memory": "128Mi",
						},
						"requests": map[string]any{
							"cpu":    "10m",
							"memory": "64Mi",
						},
					},
				},
			},
		},
	})

	return chart
}
