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
		// Each sub-service path must be routed BEFORE the catch-all centraldashboard rule.
		// The central-dashboard uses IFRAME_LINK_PREFIX='_', so clicks navigate to /_/jupyter/
		// and the iframe src is set to /jupyter/ on the same origin. Without these rules,
		// /jupyter/ would be caught by the catch-all and serve centraldashboard index.html,
		// which would show "not a valid page" error inside the iframe.
		"rules": []map[string]any{
			// Jupyter Web App — handles /jupyter/ iframe src
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/jupyter/"}},
				},
				"filters": []map[string]any{
					{"type": "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
						"set": []map[string]any{{"name": "kubeflow-userid", "value": "user@example.com"}},
					}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "jupyter-web-app-service", "port": 80, "weight": 1},
				},
			},
			// Volumes Web App
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/volumes/"}},
				},
				"filters": []map[string]any{
					{"type": "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
						"set": []map[string]any{{"name": "kubeflow-userid", "value": "user@example.com"}},
					}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "volumes-web-app-service", "port": 80, "weight": 1},
				},
			},
			// Tensorboards Web App
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/tensorboards/"}},
				},
				"filters": []map[string]any{
					{"type": "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
						"set": []map[string]any{{"name": "kubeflow-userid", "value": "user@example.com"}},
					}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "tensorboards-web-app-service", "port": 80, "weight": 1},
				},
			},
			// Katib UI
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/katib/"}},
				},
				"filters": []map[string]any{
					{"type": "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
						"set": []map[string]any{{"name": "kubeflow-userid", "value": "user@example.com"}},
					}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "katib-ui", "port": 80, "weight": 1},
				},
			},
			// Kubeflow Pipelines UI
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/pipeline/"}},
				},
				"filters": []map[string]any{
					{"type": "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
						"set": []map[string]any{{"name": "kubeflow-userid", "value": "user@example.com"}},
					}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "ml-pipeline-ui", "port": 80, "weight": 1},
				},
			},
			// Catch-all → Central Dashboard (handles /, /_/*, /api/*, etc.)
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"filters": []map[string]any{
					{"type": "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
						"set": []map[string]any{{"name": "kubeflow-userid", "value": "user@example.com"}},
					}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "centraldashboard", "port": 80, "weight": 1},
				},
			},
		},
	}))

	return chart
}
