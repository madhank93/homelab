package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewOllamaChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
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

	cdk8s.NewHelm(chart, jsii.String("ollama-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("ollama"),
		Repo:        jsii.String("https://otwld.github.io/ollama-helm"),
		Version:     jsii.String("0.79.0"),
		ReleaseName: jsii.String("ollama"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
