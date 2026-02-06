package automation

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/n8n"
)

func NewN8nChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String("automation"),
	})

	n8n.NewN8n(chart, jsii.String("n8n-release"), &n8n.N8nProps{
		ReleaseName: jsii.String("n8n"),
		Values: &n8n.N8nValues{
			AdditionalValues: &map[string]interface{}{
				// Main node configuration (UI and API)
				"main": map[string]interface{}{
					"count": 1,
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"cpu":    "100m",
							"memory": "256Mi",
						},
						"limits": map[string]interface{}{
							"cpu":    "500m",
							"memory": "512Mi",
						},
					},
					// Persistence for main node
					"persistence": map[string]interface{}{
						"enabled":    true,
						"size":       "10Gi",
						"mountPath":  "/home/node/.n8n",
						"accessMode": "ReadWriteOnce",
					},
					// No affinity field at all - let Kubernetes handle scheduling
				},

				// Service configuration
				"service": map[string]interface{}{
					"enabled": true,
					"type":    "ClusterIP",
					"port":    5678,
				},

				// Ingress configuration
				"ingress": map[string]interface{}{
					"enabled":   true,
					"className": "nginx",
					"hosts": []map[string]interface{}{
						{
							"host": "n8n.local",
							"paths": []map[string]interface{}{
								{
									"path":     "/",
									"pathType": "Prefix",
								},
							},
						},
					},
				},

				// Database configuration - PostgreSQL for production
				"db": map[string]interface{}{
					"type": "postgresdb",
				},

				// Built-in PostgreSQL using Bitnami chart
				"postgresql": map[string]interface{}{
					"enabled": true,
					"auth": map[string]interface{}{
						"database": "n8n",
						"username": "n8n",
						"password": "n8n123", // Use proper secret in production
					},
					"primary": map[string]interface{}{
						"persistence": map[string]interface{}{
							"enabled": true,
							"size":    "10Gi",
						},
					},
				},

				// Image configuration
				"image": map[string]interface{}{
					"repository": "n8nio/n8n",
					"tag":        "latest",
					"pullPolicy": "IfNotPresent",
				},

				// Security context
				"securityContext": map[string]interface{}{
					"allowPrivilegeEscalation": false,
					"capabilities": map[string]interface{}{
						"drop": []string{"ALL"},
					},
					"readOnlyRootFilesystem": false,
					"runAsNonRoot":           true,
					"runAsUser":              1000,
					"runAsGroup":             1000,
				},

				"podSecurityContext": map[string]interface{}{
					"fsGroup":             1000,
					"fsGroupChangePolicy": "OnRootMismatch",
				},

				// Service account
				"serviceAccount": map[string]interface{}{
					"create": true,
					"name":   "",
				},

				// Environment variables
				"env": map[string]interface{}{
					"N8N_HOST": "n8n.local",
					"N8N_PORT": "5678",
				},
			},
		},
	})

	return chart
}
