package observability

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

// NewArgoCDMonitorChart creates ServiceMonitors for ArgoCD components.
// ArgoCD is deployed via the platform layer (Pulumi), so this chart only
// adds the ServiceMonitors in the argocd namespace for VMAgent to discover.
func NewArgoCDMonitorChart(scope constructs.Construct, id string) cdk8s.Chart {
	namespace := "argocd"
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	components := []struct {
		name string
		port string
	}{
		// argocd-server exposes metrics on port 8083
		{"argocd-server", "metrics"},
		// argocd-repo-server exposes metrics on port 8084
		{"argocd-repo-server", "metrics"},
		// argocd-application-controller exposes metrics on port 8082
		{"argocd-application-controller", "metrics"},
	}

	for _, c := range components {
		cdk8s.NewApiObject(chart, jsii.String(c.name+"-sm"), &cdk8s.ApiObjectProps{
			ApiVersion: jsii.String("monitoring.coreos.com/v1"),
			Kind:       jsii.String("ServiceMonitor"),
			Metadata: &cdk8s.ApiObjectMetadata{
				Name:      jsii.String(c.name),
				Namespace: jsii.String(namespace),
			},
		}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
			"selector": map[string]any{
				"matchLabels": map[string]any{"app.kubernetes.io/name": c.name},
			},
			"namespaceSelector": map[string]any{
				"matchNames": []string{namespace},
			},
			"endpoints": []map[string]any{
				{"port": c.port, "path": "/metrics", "interval": "30s"},
			},
		}))
	}

	return chart
}
