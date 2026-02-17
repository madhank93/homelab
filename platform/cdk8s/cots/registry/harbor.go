package registry

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewHarborChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Create namespace
	cdk8s.NewApiObject(chart, jsii.String("harbor-namespace"), &cdk8s.ApiObjectProps{
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
					"secretsPath": "/harbor",
				},
			},
		},
		"managedSecretReference": map[string]any{
			"secretName":      "harbor-admin",
			"secretNamespace": namespace,
			"creationPolicy":  "Owner",
		},
	}

	cdk8s.NewApiObject(chart, jsii.String("harbor-infisical-secret"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets.infisical.com/v1alpha1"),
		Kind:       jsii.String("InfisicalSecret"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("harbor-secrets"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), infisicalSpec))

	values := map[string]any{
		"expose": map[string]any{
			"type": "ingress",
			"ingress": map[string]any{
				"hosts": map[string]any{
					"core": "harbor.madhan.app",
				},
				"annotations": map[string]string{
					"cert-manager.io/cluster-issuer": "letsencrypt-prod",
				},
			},
			"tls": map[string]any{
				"enabled":    true,
				"secretName": "harbor-tls",
			},
		},
		"externalURL": "https://harbor.madhan.app",
		"persistence": map[string]any{
			"enabled": true,
			"persistentVolumeClaim": map[string]any{
				"registry": map[string]any{
					"size": "50Gi",
				},
				"database": map[string]any{
					"size": "10Gi",
				},
			},
		},
		"harborAdminPassword": "",             // Not used when existingSecret is set
		"existingSecret":      "harbor-admin", // Secret created by InfisicalSecret
	}

	cdk8s.NewHelm(chart, jsii.String("harbor-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("harbor"),
		Repo:        jsii.String("https://helm.goharbor.io"),
		Version:     jsii.String("1.18.2"),
		ReleaseName: jsii.String("harbor"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
