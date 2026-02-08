package seccomp

import (
	"os"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	infisical "github.com/madhank93/homelab/cdk8s/imports/infisicalstandalone"
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

	// Create Secret for PostgreSQL password (will be sealed by CI)
	postgresSecret := cdk8s.NewApiObject(chart, jsii.String("infisical-postgresql-secret"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Secret"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("infisical-secrets"),
			Namespace: jsii.String(namespace),
		},
	})
	postgresSecret.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/stringData"), map[string]string{
		"DB_PASSWORD": infisicalDbPassword,
		"AUTH_SECRET": infisicalDbPassword, // Use same value for AUTH_SECRET (can be different in production)
	}))

	// Infisical backend + frontend + PostgreSQL + Redis
	values := map[string]any{
		"frontend": map[string]any{
			"enabled": true,
			"service": map[string]any{
				"type": "ClusterIP",
				"port": 3000,
			},
		},
		"backend": map[string]any{
			"enabled": true,
			"service": map[string]any{
				"type": "ClusterIP",
				"port": 4000,
			},
		},
		"postgresql": map[string]any{
			"enabled": true,
			"auth": map[string]any{
				"database":       "infisical",
				"username":       "infisical",
				"existingSecret": "infisical-secrets", // Use the secret we created above
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
		"redis": map[string]any{
			"enabled": true,
			"auth": map[string]any{
				"enabled": false,
			},
			"master": map[string]any{
				"persistence": map[string]any{
					"enabled": true,
					"size":    "8Gi",
				},
			},
		},
		"ingress": map[string]any{
			"enabled":  false, // Disabled - using Gateway API HTTPRoute instead
			"hostname": "infisical.madhan.app",
		},
	}

	infisical.NewInfisicalstandalone(chart, jsii.String("infisical-release"), &infisical.InfisicalstandaloneProps{
		ReleaseName: jsii.String("infisical"),
		Values:      &values,
	})

	// HTTPRoute for Infisical frontend (Gateway API + cert-manager TLS)
	httpRouteSpec := map[string]any{
		"parentRefs": []map[string]any{
			{
				"name":      "cilium-gateway",
				"namespace": "kube-system",
			},
		},
		"hostnames": []string{"infisical.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{
						"path": map[string]any{
							"type":  "PathPrefix",
							"value": "/",
						},
					},
				},
				"backendRefs": []map[string]any{
					{
						"name": "infisical-infisicalstandalone-frontend",
						"port": 3000,
					},
				},
			},
		},
	}

	cdk8s.NewApiObject(chart, jsii.String("infisical-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("infisical"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), httpRouteSpec))

	// Infisical Operator (syncs secrets from Infisical to K8s)
	cdk8s.NewHelm(chart, jsii.String("infisical-operator-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("secrets-operator"),
		Repo:        jsii.String("https://dl.cloudsmith.io/public/infisical/helm-charts/helm/charts/"),
		Version:     jsii.String("0.8.1"),
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
