package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewNvidiaGpuOperatorChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	nodeSelector := map[string]any{
		"node-role.kubernetes.io/gpu": jsii.String("true"),
	}

	values := map[string]any{
		"driver":       map[string]any{"enabled": false},
		"operator":     map[string]any{"defaultRuntime": "nvidia", "nodeSelector": nodeSelector},
		"toolkit":      map[string]any{"nodeSelector": nodeSelector},
		"devicePlugin": map[string]any{"nodeSelector": nodeSelector},
		"dcgmExporter": map[string]any{"nodeSelector": nodeSelector},
		"gfd":          map[string]any{"nodeSelector": nodeSelector},
	}

	cdk8s.NewHelm(chart, jsii.String("gpu-operator-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("gpu-operator"),
		Repo:        jsii.String("https://helm.ngc.nvidia.com/nvidia"),
		Version:     jsii.String("v24.9.1"),
		ReleaseName: jsii.String("gpu-operator"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
