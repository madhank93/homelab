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
		Version: pulumi.String("7.8.2"),
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
		OtherFields: map[string]interface{}{
			"spec": map[string]interface{}{
				"parentRefs": []map[string]interface{}{
					{
						"name":      "homelab-gateway",
						"namespace": "kube-system",
					},
				},
				"hostnames": []string{"argocd.local"},
				"rules": []map[string]interface{}{
					{
						"matches": []map[string]interface{}{
							{
								"path": map[string]interface{}{
									"type":  "PathPrefix",
									"value": "/",
								},
							},
						},
						"backendRefs": []map[string]interface{}{
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
		OtherFields: map[string]interface{}{
			"spec": map[string]interface{}{
				"parentRefs": []map[string]interface{}{
					{
						"name":      "homelab-gateway",
						"namespace": "kube-system",
					},
				},
				"hostnames": []string{"argocd.madhan.app"},
				"rules": []map[string]interface{}{
					{
						"backendRefs": []map[string]interface{}{
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

	return nil
}
