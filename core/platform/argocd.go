package platform

import (
	"github.com/madhank93/homelab/config"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// InstallArgoCD installs ArgoCD and the ApplicationSet that drives every
// workload deployment. Run `just core platform up` to apply.
//
// ArgoCD serves plain HTTP; the gateway terminates TLS using a certificate
// owned by CDK8s in workloads/observability/argocd_monitor.go.
func InstallArgoCD(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {
	argoChart := config.HelmChart("argo-cd")
	chart, err := helm.NewRelease(ctx, "argo-cd", &helm.ReleaseArgs{
		Chart:   pulumi.String("argo-cd"),
		Version: pulumi.String(argoChart.Version),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String(argoChart.Repo),
		},
		Namespace:       pulumi.String("argocd"),
		CreateNamespace: pulumi.Bool(true),
		Values: pulumi.Map{
			// Chart 10.0.0 flipped this default to true. Cilium enforces the
			// resulting policies, and ArgoCD is what would surface a breakage
			// anywhere else, so keep them off until they are reviewed on their own.
			"global": pulumi.Map{
				"networkPolicy": pulumi.Map{
					"create": pulumi.Bool(false),
				},
				// Chart 10.10.5 still pins v3.5.5. 3.5.6 stops server-side diff
				// reporting fields removed by other field managers as drift; this
				// cluster runs ServerSideApply everywhere. Drop once a chart ships it.
				"image": pulumi.Map{
					"tag": pulumi.String(config.ContainerImage("argocd").Tag),
				},
			},
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
					"type": pulumi.String("ClusterIP"), // reached only through the Gateway HTTPRoute
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
						"name":      config.C.Gateway.Name,
						"namespace": config.C.Gateway.Namespace,
					},
				},
				"hostnames": []string{"argocd.local", config.C.Host("argocd")},
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

	// 5. Create ApplicationSet to Bootstrap GitOps (watches config manifests_branch)
	// argocd_manual_sync pauses auto-sync so a release can be rolled out app by app.
	syncPolicy := map[string]any{
		// ServerSideApply=true: required for kube-prometheus-stack CRDs which
		// exceed the 262KB kubectl.kubernetes.io/last-applied-configuration limit
		"syncOptions": []string{"CreateNamespace=true", "ServerSideApply=true"},
	}
	if !config.C.ArgoCDManualSync {
		syncPolicy["automated"] = map[string]any{"prune": true, "selfHeal": true}
	}
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
							"revision": config.C.ManifestsBranch,
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
							"targetRevision": config.C.ManifestsBranch,
							"path":           "{{path}}",
						},
						"destination": map[string]any{
							"server":    "https://kubernetes.default.svc",
							"namespace": "{{path.basename}}",
						},
						"syncPolicy": syncPolicy,
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
								"group":             "apiextensions.k8s.io",
								"kind":              "CustomResourceDefinition",
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
