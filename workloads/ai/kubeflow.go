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

	// ArgoCD Application CR — tells ArgoCD to kustomize build workloads/ai/kubeflow/ from v0.1.5 branch.
	// The kustomization.yaml uses remote bases from github.com/kubeflow/manifests which CDK8s cannot inline.
	// Update targetRevision to "main" after this branch is merged.
	// Named "kubeflow-platform" (not "kubeflow") to avoid colliding with the ApplicationSet-managed
	// wrapper Application also named "kubeflow" that deploys this very directory.
	cdk8s.NewApiObject(chart, jsii.String("kubeflow-application"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("argoproj.io/v1alpha1"),
		Kind:       jsii.String("Application"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("kubeflow-platform"),
			Namespace: jsii.String("argocd"),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"project": "default",
		"source": map[string]any{
			"repoURL":        "https://github.com/madhank93/homelab.git",
			"targetRevision": "v0.1.5", // update to main after this branch is merged
			"path":           "workloads/ai/kubeflow",
		},
		"destination": map[string]any{
			"server":    "https://kubernetes.default.svc",
			"namespace": "kubeflow",
		},
		"syncPolicy": map[string]any{
			"automated": map[string]any{
				"prune":    true,
				"selfHeal": true,
			},
			"syncOptions": []string{"CreateNamespace=true", "ServerSideApply=true"},
		},
	}))

	// HTTPRoute — kubeflow.local → centraldashboard:80
	// RequestHeaderModifier injects kubeflow-userid so the dashboard identifies the user without Dex/OIDC.
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
		"hostnames": []string{"kubeflow.local"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"filters": []map[string]any{
					{
						"type": "RequestHeaderModifier",
						"requestHeaderModifier": map[string]any{
							"add": []map[string]any{
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
