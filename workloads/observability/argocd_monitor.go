package observability

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

// NewArgoCDMonitorChart creates metrics Services + ServiceMonitors for ArgoCD components.
// ArgoCD is deployed via the platform layer (Pulumi) with metrics services disabled by default.
// We create headless metrics Services here so VMAgent can scrape them.
func NewArgoCDMonitorChart(scope constructs.Construct, id string) cdk8s.Chart {
	namespace := "argocd"
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	components := []struct {
		name        string
		metricsPort float64
	}{
		{"argocd-application-controller", 8082},
		{"argocd-server", 8083},
		{"argocd-repo-server", 8084},
		{"argocd-applicationset-controller", 8085},
	}

	for _, c := range components {
		// Metrics Service — selects ArgoCD pods by app.kubernetes.io/name, exposes metrics port.
		// The ArgoCD Helm chart doesn't create metrics services by default, so we add them here.
		k8s.NewKubeService(chart, jsii.String(c.name+"-metrics-svc"), &k8s.KubeServiceProps{
			Metadata: &k8s.ObjectMeta{
				Name:      jsii.String(c.name + "-metrics"),
				Namespace: jsii.String(namespace),
				Labels: &map[string]*string{
					// ServiceMonitor selects this service by this label
					"app.kubernetes.io/name": jsii.String(c.name + "-metrics"),
				},
			},
			Spec: &k8s.ServiceSpec{
				Selector: &map[string]*string{
					// Selects the ArgoCD pods
					"app.kubernetes.io/name": jsii.String(c.name),
				},
				Ports: &[]*k8s.ServicePort{
					{
						Name:       jsii.String("metrics"),
						Port:       jsii.Number(c.metricsPort),
						TargetPort: k8s.IntOrString_FromNumber(jsii.Number(c.metricsPort)),
					},
				},
			},
		})

		cdk8s.NewApiObject(chart, jsii.String(c.name+"-sm"), &cdk8s.ApiObjectProps{
			ApiVersion: jsii.String("monitoring.coreos.com/v1"),
			Kind:       jsii.String("ServiceMonitor"),
			Metadata: &cdk8s.ApiObjectMetadata{
				Name:      jsii.String(c.name),
				Namespace: jsii.String(namespace),
			},
		}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
			"selector": map[string]any{
				"matchLabels": map[string]any{"app.kubernetes.io/name": c.name + "-metrics"},
			},
			"namespaceSelector": map[string]any{
				"matchNames": []string{namespace},
			},
			"endpoints": []map[string]any{
				{"port": "metrics", "path": "/metrics", "interval": "30s"},
			},
		}))
	}

	// Certificate for argocd.madhan.app — letsencrypt-prod DNS01 via Cloudflare.
	// ArgoCD auto-detects a secret named "argocd-server-tls" and replaces its
	// self-signed cert, fixing browser TLS warnings without public exposure.
	cdk8s.NewApiObject(chart, jsii.String("argocd-server-tls-cert"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("cert-manager.io/v1"),
		Kind:       jsii.String("Certificate"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("argocd-server-tls"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"secretName": "argocd-server-tls",
		"issuerRef": map[string]any{
			"name": "letsencrypt-prod",
			"kind": "ClusterIssuer",
		},
		"dnsNames": []string{"argocd.madhan.app"},
	}))

	return chart
}
