package observability

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/victorialogssingle"
)

func NewVictoriaLogsChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{
		"server": map[string]any{
			"enabled": true,
			"persistentVolume": map[string]any{
				"enabled": true,
				"size":    "100Gi",
			},
			"retention": "30d",
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "1000m", "memory": "1Gi"},
				"requests": map[string]any{"cpu": "200m", "memory": "256Mi"},
			},
		},
		"fluent-bit": map[string]any{
			"enabled": true,
		},
		"service": map[string]any{
			"type": "ClusterIP",
			"port": 9428,
		},
	}

	victorialogssingle.NewVictorialogssingle(chart, jsii.String("victoria-logs-release"), &victorialogssingle.VictorialogssingleProps{
		ReleaseName: jsii.String("victoria-logs"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// Gateway API HTTPRoute — routes victorialogs.madhan.app → server:9428 (/logs UI)
	cdk8s.NewApiObject(chart, jsii.String("victoria-logs-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("victoria-logs"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"victorialogs.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "victoria-logs-victoria-logs-single-server", "port": 9428, "weight": 1},
				},
			},
		},
	}))

	return chart
}
