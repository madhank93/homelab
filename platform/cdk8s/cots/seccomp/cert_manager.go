package seccomp

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewCertManagerChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{
		"installCRDs": true,
		"global": map[string]any{
			"leaderElection": map[string]any{
				"namespace": namespace,
			},
		},
	}

	// Sync Cloudflare API Token from Infisical
	// Path: /cert-manager/CLOUDFLARE_API_TOKEN -> Secret: cloudflare-api-token (key: api-token)
	infisicalSpec := map[string]any{
		"hostAPI":        "http://infisical-infisical-standalone-infisical.infisical.svc.cluster.local:8080",
		"resyncInterval": 60,
		"authentication": map[string]any{
			"serviceToken": map[string]any{
				"serviceTokenSecretReference": map[string]any{
					"secretName":      "infisical-service-token",
					"secretNamespace": "infisical",
				},
				"secretsScope": map[string]any{
					"projectSlug": "homelab-prod",
					"envSlug":     "prod",
					"secretsPath": "/cert-manager",
				},
			},
		},
		"managedSecretReference": map[string]any{
			"secretName":      "cloudflare-api-token",
			"secretNamespace": namespace,
			"creationPolicy":  "Owner",
		},
	}

	cdk8s.NewApiObject(chart, jsii.String("cloudflare-token-infisical-secret"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets.infisical.com/v1alpha1"),
		Kind:       jsii.String("InfisicalSecret"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("cloudflare-token"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), infisicalSpec))

	cdk8s.NewHelm(chart, jsii.String("cert-manager"), &cdk8s.HelmProps{
		Chart:       jsii.String("cert-manager"),
		Repo:        jsii.String("https://charts.jetstack.io"),
		Version:     jsii.String("v1.19.3"),
		ReleaseName: jsii.String("cert-manager"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// 1. Let's Encrypt ClusterIssuer (DNS-01 with Cloudflare)
	cdk8s.NewApiObject(chart, jsii.String("letsencrypt-prod-issuer"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("cert-manager.io/v1"),
		Kind:       jsii.String("ClusterIssuer"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String("letsencrypt-prod"),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
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
	}))

	// 2. Self-Signed ClusterIssuer (for internal/testing)
	cdk8s.NewApiObject(chart, jsii.String("self-signed-issuer"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("cert-manager.io/v1"),
		Kind:       jsii.String("ClusterIssuer"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String("homelab-ca"),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"selfSigned": map[string]any{},
	}))

	// 3. Wildcard Certificate for *.madhan.app (in kube-system)
	cdk8s.NewApiObject(chart, jsii.String("wildcard-certificate"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("cert-manager.io/v1"),
		Kind:       jsii.String("Certificate"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("wildcard-madhan-app"),
			Namespace: jsii.String("kube-system"), // Explicitly targeted for Gateway usage
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"secretName": "wildcard-madhan-app-tls",
		"issuerRef": map[string]any{
			"name": "letsencrypt-prod",
			"kind": "ClusterIssuer",
		},
		"dnsNames": []string{"madhan.app", "*.madhan.app"},
	}))

	return chart
}
