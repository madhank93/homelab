package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewKubeflowChart creates the HTTPRoute for the Kubeflow central dashboard and
// its sub-apps.
//
// The Kubeflow platform manifests are pre-built by CI (`kustomize build
// workloads/ai/kubeflow/`) and committed alongside this HTTPRoute in
// app/kubeflow/ on the manifests branch, so Argo CD fetches no remote kustomize
// bases at sync time.
//
// GPU notebooks in Kubeflow REQUIRE runtimeClassName: nvidia — without it the
// nvidia container hook does not fire and CUDA is inaccessible even with the
// nvidia.com/gpu resource request.
func NewKubeflowChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Every rule injects kubeflow-userid so the dashboard identifies the user
	// without Dex/OIDC in the data path.
	rule := func(prefix, svc string, rewrite bool) map[string]any {
		filters := []map[string]any{
			{"type": "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
				"set": []map[string]any{{"name": "kubeflow-userid", "value": "user@example.com"}},
			}},
		}
		if rewrite {
			filters = append(filters, map[string]any{"type": "URLRewrite", "urlRewrite": map[string]any{
				"path": map[string]any{"type": "ReplacePrefixMatch", "replacePrefixMatch": "/"},
			}})
		}
		return map[string]any{
			"matches":     []map[string]any{{"path": map[string]any{"type": "PathPrefix", "value": prefix}}},
			"filters":     filters,
			"backendRefs": []map[string]any{kit.Backend(svc, 80)},
		}
	}

	// Sub-app rules must precede the catch-all. The dashboard iframes each
	// sub-app at /<app>/; the Flask apps expect / (Istio's VirtualService
	// rewrites it), hence ReplacePrefixMatch. Katib serves natively at /katib/.
	kit.RouteRules(chart, "kubeflow-dashboard", "kubeflow", []map[string]any{
		rule("/jupyter/", "jupyter-web-app-service", true),
		rule("/volumes/", "volumes-web-app-service", true),
		rule("/tensorboards/", "tensorboards-web-app-service", true),
		rule("/katib/", "katib-ui", false),
		rule("/pipeline/", "ml-pipeline-ui", true),
		rule("/", "dashboard", false),
	})

	return chart
}
