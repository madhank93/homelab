package platform

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// InstallArgoCD installs ArgoCD via Helm and configures the GitOps bootstrap.
//
// It creates:
//   - ArgoCD Helm release (chart argo-cd, namespace argocd) in insecure mode (HTTP :80)
//   - HTTPRoute for argocd.local + argocd.madhan.app (homelab-gateway, port 80)
//   - ApplicationSet "cots-applications" watching the v0.1.6-manifests branch
//
// TLS cert for argocd.madhan.app is managed via CDK8s (workloads/observability/argocd_monitor.go).
// The gateway HTTPS listener terminates TLS using that cert before proxying to ArgoCD :80.
//
// The ApplicationSet drives all workload deployments via GitOps. Run
// `just core platform up` to apply.
func InstallArgoCD(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {
	chart, err := helm.NewRelease(ctx, "argo-cd", &helm.ReleaseArgs{
		Chart:   pulumi.String("argo-cd"),
		Version: pulumi.String("9.5.15"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://argoproj.github.io/argo-helm"),
		},
		Namespace:       pulumi.String("argocd"),
		CreateNamespace: pulumi.Bool(true),
		Values: pulumi.Map{
			"repoServer": pulumi.Map{
				// Kustomize 5.x has a hardcoded 27s git fetch timeout for remote bases.
				// KUSTOMIZE_REMOTE_FETCH_TIMEOUT overrides it so large repos (e.g. kubeflow/manifests)
				// can be fetched within CI/CD pipelines without timing out.
				"env": pulumi.Array{
					pulumi.Map{
						"name":  pulumi.String("KUSTOMIZE_REMOTE_FETCH_TIMEOUT"),
						"value": pulumi.String("180"),
					},
				},
			},
			"server": pulumi.Map{
				"extraArgs": pulumi.StringArray{},
				"service": pulumi.Map{
					"type": pulumi.String("LoadBalancer"), // Cilium L2 will assign an IP
				},
			},
			"configs": pulumi.Map{
				"params": pulumi.Map{
					"server.insecure": pulumi.Bool(true), // serve HTTP on :80; TLS terminated at Cilium gateway
					// Kubeflow manifests repo is ~400MB; default 90s is too short for git fetch from cluster.
					"reposerver.exec.timeout": pulumi.String("180s"),
				},
				"cm": pulumi.Map{
					// Kubernetes adds apiVersion, kind, volumeMode to VCT items in StatefulSets.
					// CDK8s omits them. Ignore globally so all StatefulSets stay Synced.
					"resource.customizations.ignoreDifferences.apps_StatefulSet": pulumi.String(
						"jqPathExpressions:\n" +
						"  - .spec.volumeClaimTemplates[]?.apiVersion\n" +
						"  - .spec.volumeClaimTemplates[]?.kind\n" +
						"  - .spec.volumeClaimTemplates[]?.spec.volumeMode\n",
					),
					// Kubeflow aggregationRule ClusterRoles: K8s auto-fills .rules from sub-roles;
					// manifests have rules:null. Global customization handles all ClusterRoles.
					"resource.customizations.ignoreDifferences.rbac.authorization.k8s.io_ClusterRole": pulumi.String(
						"jqPathExpressions:\n" +
						"  - .rules\n",
					),
					// Kubeflow CRDs: K8s adds .spec.conversion and .status after apply.
					// .spec.preserveUnknownFields=false is explicit in manifests but K8s drops it (default).
					"resource.customizations.ignoreDifferences.apiextensions.k8s.io_CustomResourceDefinition": pulumi.String(
						"jqPathExpressions:\n" +
						"  - .spec.conversion\n" +
						"  - .spec.preserveUnknownFields\n" +
						"  - .status\n",
					),
				},
			},
		},
	}, pulumi.Provider(k8sProvider))
	if err != nil {
		return err
	}

	// Derive the ArgoCD server service name from the Helm release name output.
	// Pulumi appends a hash to the release name (e.g. argo-cd → argo-cd-36ddc8c6);
	// chart.Name is StringPtrOutput of the actual Helm release name on the cluster.
	serverSvcName := chart.Name.ApplyT(func(name *string) string {
		if name == nil {
			return "argocd-server"
		}
		return *name + "-argocd-server"
	}).(pulumi.StringOutput)

	// HTTPRoute for ArgoCD — both LAN hostname and public hostname.
	// ArgoCD runs in insecure mode (HTTP on :80); TLS for argocd.madhan.app is
	// terminated at the Cilium gateway via the wildcard-madhan-app-tls cert.
	// The old TLSRoute (passthrough) is removed — gateway has no TLS listener on :443.
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
				"hostnames": []string{"argocd.local", "argocd.madhan.app"},
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
								"name": serverSvcName,
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

	// 5. Create ApplicationSet to Bootstrap GitOps (Watch v0.1.6-manifests)
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
							"revision": "v0.1.6-manifests", // Watch the manifests branch
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
							"targetRevision": "v0.1.6-manifests",
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
						"ignoreDifferences": []map[string]any{
							// VM operator regenerates TLS cert on every restart — ignore dynamic fields.
							{
								"kind":         "Secret",
								"name":         "victoria-metrics-victoria-metrics-operator-validation",
								"namespace":    "victoria-metrics",
								"jsonPointers": []string{"/data"},
							},
							{
								"group":             "admissionregistration.k8s.io",
								"kind":              "ValidatingWebhookConfiguration",
								"name":              "victoria-metrics-victoria-metrics-operator-admission",
								"jqPathExpressions": []string{".webhooks[].clientConfig.caBundle"},
							},
							// Kubeflow CRDs and ClusterRoles: Kubernetes mutates fields not in manifests.
							{
								"group": "apiextensions.k8s.io",
								"kind": "CustomResourceDefinition",
								"jqPathExpressions": []string{".spec.conversion", ".status"},
							},
							{
								// aggregationRule ClusterRoles: K8s auto-fills .rules from sub-roles;
								// manifest has rules:null — use jsonPointers which handles null reliably.
								"kind":         "ClusterRole",
								"jsonPointers": []string{"/metadata/annotations", "/rules"},
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
