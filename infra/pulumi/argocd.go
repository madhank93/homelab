package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func InstallArgoCD(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {

	// Define ArgoCD Helm Chart
	chart, err := helm.NewRelease(ctx, "argo-cd", &helm.ReleaseArgs{
		Chart:   pulumi.String("argo-cd"),
		Version: pulumi.String("9.4.2"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://argoproj.github.io/argo-helm"),
		},
		Namespace:       pulumi.String("argocd"),
		CreateNamespace: pulumi.Bool(true),
		Values: pulumi.Map{
			"server": pulumi.Map{
				"extraArgs": pulumi.StringArray{},
				"service": pulumi.Map{
					"type": pulumi.String("LoadBalancer"), // Cilium L2 will assign an IP
				},
			},
			"configs": pulumi.Map{
				"params": pulumi.Map{
					"server.insecure": pulumi.Bool(false),
				},
			},
		},
	}, pulumi.Provider(k8sProvider))

	// Define HTTPRoute for ArgoCD (Gateway API)
	_, err = apiextensions.NewCustomResource(ctx, "argocd-httproute", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("gateway.networking.k8s.io/v1"),
		Kind:       pulumi.String("HTTPRoute"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("argocd-server-route"),
			Namespace: pulumi.String("argocd"),
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"parentRefs": []map[string]any{
					{
						"name":      "homelab-gateway",
						"namespace": "kube-system",
					},
				},
				"hostnames": []string{"argocd.local"},
				"rules": []map[string]any{
					{
						"matches": []map[string]any{
							{
								"path": map[string]any{
									"type":  "PathPrefix",
									"value": "/",
								},
							},
						},
						"backendRefs": []map[string]any{
							{
								"name": "argocd-server",
								"port": 80,
							},
						},
					},
				},
			},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{chart}))
	if err != nil {
		return err
	}

	// Create TLSRoute for ArgoCD
	_, err = apiextensions.NewCustomResource(ctx, "argocd-tlsmetadata", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("gateway.networking.k8s.io/v1alpha2"),
		Kind:       pulumi.String("TLSRoute"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("argocd-route"),
			Namespace: pulumi.String("argocd"),
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"parentRefs": []map[string]any{
					{
						"name":      "homelab-gateway",
						"namespace": "kube-system",
					},
				},
				"hostnames": []string{"argocd.madhan.app"},
				"rules": []map[string]any{
					{
						"backendRefs": []map[string]any{
							{
								"name": "argo-cd-964152f1-argocd-server",
								"port": 443,
							},
						},
					},
				},
			},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{chart}))
	if err != nil {
		return err
	}

	// 5. Create ApplicationSet to Bootstrap GitOps (Watch v0.1.5-manifests)
	_, err = apiextensions.NewCustomResource(ctx, "bootstrap-appset", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("argoproj.io/v1alpha1"),
		Kind:       pulumi.String("ApplicationSet"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("cots-applications"),
			Namespace: pulumi.String("argocd"),
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"generators": []map[string]any{
					{
						"git": map[string]any{
							"repoURL":  "https://github.com/madhank93/homelab.git",
							"revision": "v0.1.5-manifests", // Watch the manifests branch
							"directories": []map[string]any{
								{"path": "*"}, // Apps are at the root of the manifests branch
							},
						},
					},
				},
				"template": map[string]any{
					"metadata": map[string]any{
						"name": "{{path.basename}}",
					},
					"spec": map[string]any{
						"project": "default",
						"source": map[string]any{
							"repoURL":        "https://github.com/madhank93/homelab.git",
							"targetRevision": "v0.1.5-manifests",
							"path":           "{{path}}",
						},
						"destination": map[string]any{
							"server":    "https://kubernetes.default.svc",
							"namespace": "{{path.basename}}",
						},
						"syncPolicy": map[string]any{
							"automated": map[string]any{
								"prune":    true,
								"selfHeal": true,
							},
							// ServerSideApply=true: required for kube-prometheus-stack CRDs which
							// exceed the 262KB kubectl.kubernetes.io/last-applied-configuration limit
							"syncOptions": []string{"CreateNamespace=true", "ServerSideApply=true"},
						},
						// Infisical CRD schema omits projectSlug from secretsScope —
						// skip structured merge diff for InfisicalSecret spec to avoid
						// "field not declared in schema" comparison errors
						"ignoreDifferences": []map[string]any{
							{
								"group":        "secrets.infisical.com",
								"kind":         "InfisicalSecret",
								"jsonPointers": []string{"/spec"},
							},
						},
					},
				},
			},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{chart}))
	if err != nil {
		return err
	}

	return nil
}
