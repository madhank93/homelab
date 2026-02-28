package platform

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// InstallCertManager installs cert-manager and configures issuers + wildcard certificate.
// Prerequisite: cert-manager/cloudflare-api-token Secret must exist before the
// letsencrypt-prod ClusterIssuer can complete DNS-01 challenges. Create it with:
//
//	just create-secrets
func InstallCertManager(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {

	// Install cert-manager via Helm (CRDs bundled via installCRDs: true)
	chart, err := helm.NewRelease(ctx, "cert-manager", &helm.ReleaseArgs{
		Name:    pulumi.String("cert-manager"),
		Chart:   pulumi.String("cert-manager"),
		Version: pulumi.String("v1.19.3"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://charts.jetstack.io"),
		},
		Namespace:       pulumi.String("cert-manager"),
		CreateNamespace: pulumi.Bool(true),
		Values: pulumi.Map{
			"installCRDs": pulumi.Bool(true),
			"global": pulumi.Map{
				"leaderElection": pulumi.Map{
					"namespace": pulumi.String("cert-manager"),
				},
			},
			"resources": pulumi.Map{
				"limits":   pulumi.Map{"cpu": pulumi.String("200m"), "memory": pulumi.String("256Mi")},
				"requests": pulumi.Map{"cpu": pulumi.String("50m"), "memory": pulumi.String("64Mi")},
			},
			"webhook": pulumi.Map{
				"resources": pulumi.Map{
					"limits":   pulumi.Map{"cpu": pulumi.String("100m"), "memory": pulumi.String("128Mi")},
					"requests": pulumi.Map{"cpu": pulumi.String("50m"), "memory": pulumi.String("64Mi")},
				},
			},
			"cainjector": pulumi.Map{
				"resources": pulumi.Map{
					"limits":   pulumi.Map{"cpu": pulumi.String("100m"), "memory": pulumi.String("128Mi")},
					"requests": pulumi.Map{"cpu": pulumi.String("50m"), "memory": pulumi.String("64Mi")},
				},
			},
		},
	}, pulumi.Provider(k8sProvider))
	if err != nil {
		return err
	}

	// Let's Encrypt ClusterIssuer (DNS-01 via Cloudflare)
	// Requires cert-manager/cloudflare-api-token Secret — created by create-bootstrap-secrets.sh
	_, err = apiextensions.NewCustomResource(ctx, "letsencrypt-prod-issuer", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("ClusterIssuer"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("letsencrypt-prod"),
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"acme": map[string]any{
					"server": "https://acme-v02.api.letsencrypt.org/directory",
					"email":  "madhankumaravelu93@gmail.com",
					"privateKeySecretRef": map[string]any{
						"name": "letsencrypt-prod-key",
					},
					"solvers": []map[string]any{
						{
							"dns01": map[string]any{
								"cloudflare": map[string]any{
									"email": "madhankumaravelu93@gmail.com",
									"apiTokenSecretRef": map[string]any{
										"name": "cloudflare-api-token",
										"key":  "CLOUDFLARE_API_TOKEN",
									},
								},
							},
							"selector": map[string]any{
								"dnsZones": []string{"madhan.app"},
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

	// Self-Signed ClusterIssuer (for internal / testing use)
	_, err = apiextensions.NewCustomResource(ctx, "self-signed-issuer", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("ClusterIssuer"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("homelab-ca"),
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"selfSigned": map[string]any{},
			},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{chart}))
	if err != nil {
		return err
	}

	// Wildcard certificate for *.madhan.app
	// Stored in kube-system so the Gateway can reference it across namespaces.
	// The HTTPS listener in cilium.go can be re-enabled once this Secret exists.
	_, err = apiextensions.NewCustomResource(ctx, "wildcard-certificate", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("Certificate"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("wildcard-madhan-app"),
			Namespace: pulumi.String("kube-system"),
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"secretName": "wildcard-madhan-app-tls",
				"issuerRef": map[string]any{
					"name": "letsencrypt-prod",
					"kind": "ClusterIssuer",
				},
				"dnsNames": []string{"madhan.app", "*.madhan.app"},
			},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{chart}))

	return err
}
