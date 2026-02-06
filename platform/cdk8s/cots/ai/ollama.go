package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/ollama"
)

func NewOllamaChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String("ai"),
	})

	gpuNodeSelector := map[string]any{
		"node-role.kubernetes.io/gpu": jsii.String("true"),
	}

	values := map[string]any{
		"replicaCount": 1,
		"image": map[string]any{
			"repository": "ollama/ollama",
			"tag":        "latest",
		},
		"resources": map[string]any{
			"limits": map[string]any{
				"nvidia.com/gpu": 1,
				"memory":         "8Gi",
				"cpu":            "2000m",
			},
			"requests": map[string]any{
				"memory": "4Gi",
				"cpu":    "1000m",
			},
		},
		"persistence": map[string]any{
			"enabled": true,
			"size":    "50Gi",
		},
		"service": map[string]any{
			"type": "ClusterIP",
			"port": 11434,
		},
		"nodeSelector": gpuNodeSelector,
	}

	ollama.NewOllama(chart, jsii.String("ollama-release"), &ollama.OllamaProps{
		ReleaseName: jsii.String("ollama"),
		Values:      &values,
	})

	return chart
}
