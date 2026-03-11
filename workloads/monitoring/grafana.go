package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/grafana"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

func NewGrafanaChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("monitoring-namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
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

	values := map[string]interface{}{
		"podAnnotations": map[string]interface{}{
			"reloader.stakater.com/auto": "true",
		},
		"datasources": map[string]interface{}{
			"datasources.yaml": map[string]interface{}{
				"apiVersion": 1,
				"datasources": []map[string]interface{}{
					{
						"name":      "VictoriaMetrics",
						"type":      "prometheus",
						"url":       "http://victoria-metrics-victoria-metrics-cluster-vmselect.victoria-metrics.svc.cluster.local:8481/select/0/prometheus",
						"access":    "proxy",
						"isDefault": true,
						"jsonData": map[string]interface{}{
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
		// Admin credentials: user set directly in env; password read from CSI-mounted file.
		// The Grafana chart's auto-generated admin Secret is NOT used to avoid synth-time
		// random value churn. GF_SECURITY_ADMIN_PASSWORD__FILE takes precedence over the Secret.
		"env": map[string]interface{}{
			"GF_SECURITY_ADMIN_USER":           "admin",
			"GF_SECURITY_ADMIN_PASSWORD__FILE": "/mnt/secrets/ADMIN_PASSWORD",
		},
		"resources": map[string]interface{}{
			"limits":   map[string]interface{}{"cpu": "500m", "memory": "512Mi"},
			"requests": map[string]interface{}{"cpu": "100m", "memory": "128Mi"},
		},
		"persistence": map[string]interface{}{
			"enabled": true,
			"size":    "10Gi",
		},
		"service": map[string]interface{}{
			"type": "ClusterIP",
			"port": 3000,
		},
		"ingress": map[string]interface{}{"enabled": false},
		"extraVolumes": []map[string]interface{}{
			{
				"name": "openbao-secrets",
				"csi": map[string]interface{}{
					"driver":   "secrets-store.csi.k8s.io",
					"readOnly": true,
					"volumeAttributes": map[string]interface{}{
						"secretProviderClass": "grafana-secrets",
					},
				},
			},
		},
		"extraVolumeMounts": []map[string]interface{}{
			{
				"name":      "openbao-secrets",
				"mountPath": "/mnt/secrets",
				"readOnly":  true,
			},
		},
	}

	grafana.NewGrafana(chart, jsii.String("grafana-release"), &grafana.GrafanaProps{
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
