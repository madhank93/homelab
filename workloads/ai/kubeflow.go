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
		// Sub-service rules must come before the catch-all centraldashboard rule.
		// The central-dashboard uses IFRAME_LINK_PREFIX='_': clicking "Notebooks" sets the
		// browser URL to /_/jupyter/ and the iframe src to /jupyter/ (same origin).
		//
		// Each sub-app is a Python Flask SPA with APP_PREFIX=/jupyter etc. In Istio, the
		// VirtualService rewrites /jupyter/ → / before forwarding to the Flask app, so Flask
		// sees /static/runtime.js instead of /jupyter/static/runtime.js. We replicate that
		// here with URLRewrite ReplacePrefixMatch so static assets and API routes all resolve.
		"rules": []map[string]any{
			// Jupyter Web App (/jupyter/ → /)
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/jupyter/"}},
				},
				"filters": []map[string]any{
					{"type": "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
						"set": []map[string]any{{"name": "kubeflow-userid", "value": "user@example.com"}},
					}},
					{"type": "URLRewrite", "urlRewrite": map[string]any{
						"path": map[string]any{"type": "ReplacePrefixMatch", "replacePrefixMatch": "/"},
					}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "jupyter-web-app-service", "port": 80, "weight": 1},
				},
			},
			// Volumes Web App (/volumes/ → /)
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/volumes/"}},
				},
				"filters": []map[string]any{
					{"type": "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
						"set": []map[string]any{{"name": "kubeflow-userid", "value": "user@example.com"}},
					}},
					{"type": "URLRewrite", "urlRewrite": map[string]any{
						"path": map[string]any{"type": "ReplacePrefixMatch", "replacePrefixMatch": "/"},
					}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "volumes-web-app-service", "port": 80, "weight": 1},
				},
			},
			// Tensorboards Web App (/tensorboards/ → /)
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/tensorboards/"}},
				},
				"filters": []map[string]any{
					{"type": "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
						"set": []map[string]any{{"name": "kubeflow-userid", "value": "user@example.com"}},
					}},
					{"type": "URLRewrite", "urlRewrite": map[string]any{
						"path": map[string]any{"type": "ReplacePrefixMatch", "replacePrefixMatch": "/"},
					}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "tensorboards-web-app-service", "port": 80, "weight": 1},
				},
			},
			// Katib UI — prefix is baked into the Go binary, serves natively at /katib/. No URLRewrite.
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
			// Kubeflow Pipelines UI (/pipeline/ → /)
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/pipeline/"}},
				},
				"filters": []map[string]any{
					{"type": "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
						"set": []map[string]any{{"name": "kubeflow-userid", "value": "user@example.com"}},
					}},
					{"type": "URLRewrite", "urlRewrite": map[string]any{
						"path": map[string]any{"type": "ReplacePrefixMatch", "replacePrefixMatch": "/"},
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
