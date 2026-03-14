package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewKubeflowChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// HTTPRoute — kubeflow.local → centraldashboard:80
	// RequestHeaderModifier injects kubeflow-userid so the dashboard identifies the user without Dex/OIDC.
	// The Kubeflow platform manifests are pre-built by CI (kustomize build workloads/ai/kubeflow/)
	// and committed alongside this HTTPRoute in app/kubeflow/ on the manifests branch.
	// ArgoCD's ApplicationSet deploys app/kubeflow/ directly — no kustomize remote base fetching at sync time.
	cdk8s.NewApiObject(chart, jsii.String("kubeflow-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("kubeflow-dashboard"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"kubeflow.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"filters": []map[string]any{
					{
						"type": "RequestHeaderModifier",
						"requestHeaderModifier": map[string]any{
							// set replaces (not appends) to avoid duplicate header values
							"set": []map[string]any{
								{"name": "kubeflow-userid", "value": "user@example.com"},
							},
						},
					},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "centraldashboard", "port": 80, "weight": 1},
				},
			},
		},
	}))

	return chart
}
