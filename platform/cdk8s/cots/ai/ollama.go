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

	// nvidia.com/gpu.present is set by NFD — matches the selector used by the GPU operator.
	gpuNodeSelector := map[string]any{
		"nvidia.com/gpu.present": jsii.String("true"),
	}

	values := map[string]any{
		"replicaCount": 1,
		"image": map[string]any{
			"repository": "ollama/ollama",
			"tag":        "0.17.0",
		},
		"resources": map[string]any{
			"limits": map[string]any{
				"nvidia.com/gpu": 1,
				// memory here is host RAM (cgroup limit), NOT GPU VRAM.
				// GPU VRAM (16GB) is fully available via nvidia.com/gpu: 1.
				// Worker4 has 6GB host RAM total; keep limit below that.
				"memory": "4Gi",
				"cpu":    "4000m",
			},
			"requests": map[string]any{
				"memory": "2Gi",
				"cpu":    "1000m",
			},
		},
		"persistence": map[string]any{
			"enabled": true,
			"size":    "100Gi",
		},
		"service": map[string]any{
			"type": "ClusterIP",
			"port": 11434,
		},
		"nodeSelector": gpuNodeSelector,
		// runtimeClassName=nvidia: routes the pod through nvidia-container-runtime
		// (installed by Talos nvidia-container-toolkit-production extension).
		// The runtime injects GPU devices using Talos-aware paths (/usr/local/glibc/usr/lib/)
		// based on NVIDIA_VISIBLE_DEVICES set by the device plugin (envvars mode).
		"runtimeClassName": "nvidia",
		"extraEnv": []map[string]any{
			{"name": "NVIDIA_VISIBLE_DEVICES", "value": "all"},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("ollama-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("ollama"),
		Repo:        jsii.String("https://otwld.github.io/ollama-helm"),
		Version:     jsii.String("1.41.0"),
		ReleaseName: jsii.String("ollama"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
