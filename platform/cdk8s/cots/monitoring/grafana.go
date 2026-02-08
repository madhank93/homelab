package monitoring

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewGrafanaChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Create namespace
	cdk8s.NewApiObject(chart, jsii.String("monitoring-namespace"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Namespace"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String(namespace),
		},
	})

	// Create InfisicalSecret CRD to sync secrets from Infisical
	infisicalSpec := map[string]any{
		"hostAPI":        "http://infisical-infisicalstandalone-backend.infisical.svc.cluster.local:4000",
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
					"secretsPath": "/grafana",
				},
			},
		},
		"managedSecretReference": map[string]any{
			"secretName":      "grafana-admin",
			"secretNamespace": namespace,
			"creationPolicy":  "Owner",
		},
	}

	cdk8s.NewApiObject(chart, jsii.String("grafana-infisical-secret"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets.infisical.com/v1alpha1"),
		Kind:       jsii.String("InfisicalSecret"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("grafana-secrets"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), infisicalSpec))

	// Grafana Helm chart configuration
	values := map[string]any{
		"admin": map[string]any{
			"existingSecret": "grafana-admin",  // Secret created by InfisicalSecret
			"passwordKey":    "ADMIN_PASSWORD", // Key from Infisical
		},
		"persistence": map[string]any{
			"enabled": true,
			"size":    "10Gi",
		},
		"service": map[string]any{
			"type": "ClusterIP",
			"port": 3000,
		},
		"ingress": map[string]any{
			"enabled": true,
			"hosts":   []string{"grafana.madhan.app"},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("grafana-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("grafana"),
		Repo:        jsii.String("https://grafana.github.io/helm-charts"),
		Version:     jsii.String("10.5.15"),
		ReleaseName: jsii.String("grafana"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
