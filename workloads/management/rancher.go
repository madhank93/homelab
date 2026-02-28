package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewRancherChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Create namespace
	cdk8s.NewApiObject(chart, jsii.String("rancher-namespace"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Namespace"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String(namespace),
		},
	})

	// Create InfisicalSecret CRD
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
					"secretsPath": "/rancher",
				},
			},
		},
		"managedSecretReference": map[string]any{
			"secretName":      "rancher-bootstrap",
			"secretNamespace": namespace,
			"creationPolicy":  "Owner",
		},
	}

	cdk8s.NewApiObject(chart, jsii.String("rancher-infisical-secret"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets.infisical.com/v1alpha1"),
		Kind:       jsii.String("InfisicalSecret"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("rancher-secrets"),
			Namespace: jsii.String(namespace),
			// ServerSideApply=false: Infisical CRD schema omits projectSlug from
			// serviceToken.secretsScope, causing SSA schema validation to fail.
			// Use client-side apply for this resource only.
			Annotations: &map[string]*string{
				"argocd.argoproj.io/sync-options": jsii.String("ServerSideApply=false"),
			},
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), infisicalSpec))

	values := map[string]any{
		"agentTLSMode": "system-store",
		"auditLog": map[string]any{
			"level":       0,
			"destination": "sidecar",
		},
		// Ingress disabled — traffic routed via Gateway API HTTPRoute below
		// Rancher's built-in ingress required an Nginx controller; removed in favour of homelab-gateway
		"ingress": map[string]any{
			"enabled": false,
		},
		"service": map[string]any{
			"type":        "ClusterIP",
			"disableHttp": false,
		},
		"hostname":                   "rancher.madhan.app", // Updated to real domain
		"bootstrapPassword":          "",                   // Not used when secret exists
		"existingBootstrapPassword":  "rancher-bootstrap",  // Secret created by InfisicalSecret
		"bootstrapPasswordSecretKey": "BOOTSTRAP_PASSWORD",
		"replicas":                   3,
		"resources": map[string]any{
			"limits": map[string]any{
				"memory": "2Gi",
				"cpu":    "1000m",
			},
			"requests": map[string]any{
				"memory": "1Gi",
				"cpu":    "500m",
			},
		},
		"antiAffinity": "preferred",
	}

	cdk8s.NewHelm(chart, jsii.String("rancher-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("rancher"),
		Repo:        jsii.String("https://releases.rancher.com/server-charts/stable"),
		ReleaseName: jsii.String("rancher"),
		Namespace:   jsii.String(namespace),
		Version:     jsii.String("2.13.2"),
		HelmFlags:   &[]*string{jsii.String("--kube-version"), jsii.String("1.30.0")},
		Values:      &values,
	})

	// Gateway API HTTPRoute — routes rancher.madhan.app → rancher:80
	cdk8s.NewApiObject(chart, jsii.String("rancher-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("rancher"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"rancher.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"name": "rancher", "port": 80},
				},
			},
		},
	}))

	return chart
}
