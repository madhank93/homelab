package platform

import (
	"fmt"

	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/core/internal/cfg"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// InstallArgoCD installs Argo CD with Authentik SSO, its HTTPRoute, and the
// ApplicationSet that creates one Application per top-level directory of the
// manifests branch. Argo CD serves plain HTTP; the Gateway terminates TLS.
func InstallArgoCD(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {
	oidcSecret := cfg.K.String("ARGOCD_OAUTH_CLIENT_SECRET")
	if oidcSecret == "" {
		return fmt.Errorf("ARGOCD_OAUTH_CLIENT_SECRET not found in .env")
	}

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
			// The chart creates NetworkPolicies by default. Cilium enforces them, and
			// Argo CD is what would surface a breakage anywhere else, so they stay off
			// until reviewed on their own.
			"global": pulumi.Map{
				"networkPolicy": pulumi.Map{
					"create": pulumi.Bool(false),
				},
				// Overrides the chart's default image: v3.5.6+ stops server-side diff
				// reporting fields removed by other field managers as drift, and this
				// cluster uses ServerSideApply everywhere. Drop once the chart default
				// reaches it.
				"image": pulumi.Map{
					"tag": pulumi.String(config.ContainerImage("argocd").Tag),
				},
			},
			"repoServer": pulumi.Map{
				// Kustomize's default git fetch timeout for remote bases is too short
				// for large repos such as kubeflow/manifests.
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
				"rbac": pulumi.Map{
					// SSO users get nothing unless they are in grafana-admins.
					"policy.default": pulumi.String(""),
					"policy.csv":     pulumi.String("g, grafana-admins, role:admin\n"),
					"scopes":         pulumi.String("[groups]"),
				},
				"params": pulumi.Map{
					"server.insecure": pulumi.Bool(true), // serve HTTP on :80; TLS terminated at Cilium gateway
					// The Kubeflow manifests repo is too large for the default 90s.
					"reposerver.exec.timeout": pulumi.String("180s"),
				},
				"cm": pulumi.Map{
					"url": pulumi.String("https://" + config.C.Host("argocd")),
					// SSO through Authentik; the local admin login stays as the fallback.
					// $argocd-oidc:clientSecret reads the Secret created below.
					"oidc.config": pulumi.String(
						"name: Authentik\n" +
							"issuer: https://" + config.C.Host("auth") + "/application/o/argocd/\n" +
							"clientID: argocd-homelab\n" +
							"clientSecret: $argocd-oidc:clientSecret\n" +
							"requestedScopes: [openid, profile, email, groups]\n",
					),
					// Kubernetes adds apiVersion, kind and volumeMode to StatefulSet
					// volumeClaimTemplates, which CDK8s omits. Set globally because
					// per-Application jqPathExpressions are not applied to live state
					// under server-side diff.
					"resource.customizations.ignoreDifferences.apps_StatefulSet": pulumi.String(
						"jqPathExpressions:\n" +
							"  - .spec.volumeClaimTemplates[]?.apiVersion\n" +
							"  - .spec.volumeClaimTemplates[]?.kind\n" +
							"  - .spec.volumeClaimTemplates[]?.spec.volumeMode\n",
					),
					// Kubernetes fills .rules and annotations on aggregated ClusterRoles.
					// jsonPointers, unlike jq, tolerates Kubeflow's rules: null.
					"resource.customizations.ignoreDifferences.rbac.authorization.k8s.io_ClusterRole": pulumi.String(
						"jsonPointers:\n" +
							"  - /metadata/annotations\n" +
							"  - /rules\n",
					),
					// Kubernetes adds .spec.conversion and .status to CRDs, and drops
					// preserveUnknownFields: false, which Kubeflow sets explicitly.
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

	// Argo CD deploys OpenBao, so its OIDC client secret cannot come from there.
	// The part-of label is what lets argocd-cm reference this Secret.
	_, err = corev1.NewSecret(ctx, "argocd-oidc", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("argocd-oidc"),
			Namespace: pulumi.String("argocd"),
			Labels:    pulumi.StringMap{"app.kubernetes.io/part-of": pulumi.String("argocd")},
		},
		StringData: pulumi.StringMap{"clientSecret": pulumi.ToSecret(pulumi.String(oidcSecret)).(pulumi.StringOutput)},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{chart}))
	if err != nil {
		return err
	}

	// The release has no fixed name, so Pulumi suffixes it (argo-cd-<hash>) and
	// the server Service name must be derived from it.
	serverSvcName := chart.Name.ApplyT(func(name *string) string {
		if name == nil {
			return "argocd-server"
		}
		return *name + "-argocd-server"
	}).(pulumi.StringOutput)

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

	// argocd_manual_sync turns off auto-sync so a release can be rolled out app
	// by app.
	syncPolicy := map[string]any{
		// Server-side apply avoids the last-applied-configuration annotation,
		// which large CRDs overflow (256 KiB annotation limit).
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
								{"path": "*"}, // one directory per app at the branch root
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
							// The VictoriaMetrics operator regenerates its webhook cert on every restart.
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
