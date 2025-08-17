package cots

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	nvidia "github.com/madhank93/homelab/cdk8s/imports/gpuoperator"
)

func NewNvidiaGpuOperatorChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String("gpu-operator"),
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

	nvidia.NewGpuoperator(chart, jsii.String("gpu-operator-release"), &nvidia.GpuoperatorProps{
		ReleaseName: jsii.String("gpu-operator"),
		Values:      &values,
	})

	return chart
}
