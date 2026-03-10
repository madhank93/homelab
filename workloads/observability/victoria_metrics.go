package observability

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/victoriametricscluster"
)

func NewVictoriaMetricsChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{
		"server": map[string]any{
			"enabled": true,
			"persistentVolume": map[string]any{
				"enabled": true,
				"size":    "50Gi",
			},
			"retention": "30d",
		},
		"vmselect": map[string]any{
			"enabled":      true,
			"replicaCount": 1,
			"resources": map[string]any{
				"limits": map[string]any{
					"memory": "1Gi",
					"cpu":    "500m",
				},
			},
		},
		"vminsert": map[string]any{
			"enabled":      true,
			"replicaCount": 1,
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
		},
		"vmstorage": map[string]any{
			"enabled":      true,
			"replicaCount": 1,
			"persistentVolume": map[string]any{
				"enabled": true,
				"size":    "100Gi",
			},
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "1000m", "memory": "1Gi"},
				"requests": map[string]any{"cpu": "200m", "memory": "256Mi"},
			},
		},
	}

	victoriametricscluster.NewVictoriametricscluster(chart, jsii.String("victoria-metrics-release"), &victoriametricscluster.VictoriametricsclusterProps{
		ReleaseName: jsii.String("victoria-metrics"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// Gateway API HTTPRoute — routes vmselect.madhan.app → vmselect:8481 (/vmui/ web UI)
	cdk8s.NewApiObject(chart, jsii.String("victoria-metrics-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("victoria-metrics"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"vmselect.madhan.app"},
		"rules": []map[string]any{
			{
				// Redirect bare root to the vmui path
				"matches": []map[string]any{
					{"path": map[string]any{"type": "Exact", "value": "/"}},
				},
				"filters": []map[string]any{
					{
						"type": "RequestRedirect",
						"requestRedirect": map[string]any{
							"path": map[string]any{
								"type":            "ReplaceFullPath",
								"replaceFullPath": "/select/0/vmui/",
							},
							"statusCode": 302,
						},
					},
				},
			},
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "victoria-metrics-victoria-metrics-cluster-vmselect", "port": 8481, "weight": 1},
				},
			},
		},
	}))

	return chart
}
