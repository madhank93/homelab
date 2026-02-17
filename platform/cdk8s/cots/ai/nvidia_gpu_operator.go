package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/k8s"
)

func NewNvidiaGpuOperatorChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
		},
	})

	nodeSelector := map[string]any{
		"node-role.kubernetes.io/gpu": jsii.String("true"),
	}

	values := map[string]any{
		"driver": map[string]any{"enabled": false},
		"operator": map[string]any{
			"defaultRuntime": "nvidia",
			"nodeSelector":   nodeSelector,
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
		},
		"toolkit":      map[string]any{"nodeSelector": nodeSelector},
		"devicePlugin": map[string]any{"nodeSelector": nodeSelector},
		"dcgmExporter": map[string]any{"nodeSelector": nodeSelector},
		"gfd":          map[string]any{"nodeSelector": nodeSelector},
	}

	// CRDs
	cdk8s.NewInclude(chart, jsii.String("nvidia-cluster-policy-crd"), &cdk8s.IncludeProps{
		Url: jsii.String("https://raw.githubusercontent.com/NVIDIA/gpu-operator/v25.10.1/deployments/gpu-operator/crds/nvidia.com_clusterpolicies.yaml"),
	})
	cdk8s.NewInclude(chart, jsii.String("nvidia-driver-crd"), &cdk8s.IncludeProps{
		Url: jsii.String("https://raw.githubusercontent.com/NVIDIA/gpu-operator/v25.10.1/deployments/gpu-operator/crds/nvidia.com_nvidiadrivers.yaml"),
	})

	cdk8s.NewHelm(chart, jsii.String("gpu-operator-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("gpu-operator"),
		Repo:        jsii.String("https://helm.ngc.nvidia.com/nvidia"),
		Version:     jsii.String("v25.10.1"),
		ReleaseName: jsii.String("gpu-operator"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
