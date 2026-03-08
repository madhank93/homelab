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

	// SecretProviderClass — Pattern A (file-only, no secretObjects sync).
	// Mounts ADMIN_PASSWORD from OpenBao at /mnt/secrets/ADMIN_PASSWORD.
	// Grafana reads it via GF_SECURITY_ADMIN_PASSWORD__FILE env var.
	cdk8s.NewApiObject(chart, jsii.String("grafana-spc"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets-store.csi.x-k8s.io/v1"),
		Kind:       jsii.String("SecretProviderClass"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("grafana-secrets"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"provider": "openbao",
		"parameters": map[string]any{
			"vaultAddress": "http://openbao.openbao.svc.cluster.local:8200",
			"roleName":     "grafana",
			"objects": `- objectName: "ADMIN_PASSWORD"
  secretPath: "secret/data/grafana"
  secretKey: "ADMIN_PASSWORD"`,
		},
	}))

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
		// Admin password read from file — avoids storing password in a k8s Secret.
		// File is mounted from OpenBao via the CSI volume below.
		"admin": map[string]any{
			"userKey":     "admin",
			"passwordKey": "admin",
		},
		"env": map[string]any{
			"GF_SECURITY_ADMIN_PASSWORD__FILE": "/mnt/secrets/ADMIN_PASSWORD",
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
		// CSI volume mounts OpenBao secrets as files at /mnt/secrets/
		// Required to trigger SecretProviderClass to fetch secrets from OpenBao.
		"extraVolumes": []map[string]any{
			{
				"name": "openbao-secrets",
				"csi": map[string]any{
					"driver":   "secrets-store.csi.k8s.io",
					"readOnly": true,
					"volumeAttributes": map[string]any{
						"secretProviderClass": "grafana-secrets",
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
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"grafana.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "grafana", "port": 3000, "weight": 1},
				},
			},
		},
	}))

	return chart
}
