package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewHeadlampChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	cdk8s.NewHelm(chart, jsii.String("headlamp-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("headlamp"),
		Repo:        jsii.String("https://kubernetes-sigs.github.io/headlamp/"),
		ReleaseName: jsii.String("headlamp"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
			// Ingress disabled — traffic routed via Gateway API HTTPRoute below
			"ingress": map[string]any{
				"enabled": false,
			},
		},
	})

	// Gateway API HTTPRoute — routes headlamp.madhan.app → headlamp:4466
	cdk8s.NewApiObject(chart, jsii.String("headlamp-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("headlamp"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"headlamp.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"name": "headlamp", "port": 4466},
				},
			},
		},
	}))

	return chart
}
