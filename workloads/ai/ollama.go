package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/ollama"
)

func NewOllamaChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// nvidia.com/gpu.present is set by NFD — matches the selector used by the GPU operator.
	gpuNodeSelector := map[string]any{
		"nvidia.com/gpu.present": "true",
	}

	ollama.NewOllama(chart, jsii.String("ollama-release"), &ollama.OllamaProps{
		ReleaseName: jsii.String("ollama"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
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
			"tolerations": []map[string]any{
				{"key": "dedicated", "operator": "Equal", "value": "ai", "effect": "NoSchedule"},
			},
			// runtimeClassName=nvidia: routes the pod through nvidia-container-runtime
			// (installed by Talos nvidia-container-toolkit-production extension).
			"runtimeClassName": "nvidia",
			"extraEnv": []map[string]any{
				{"name": "NVIDIA_VISIBLE_DEVICES", "value": "all"},
			},
		},
	})

	// Gateway API HTTPRoute — routes ollama.madhan.app → ollama:11434
	cdk8s.NewApiObject(chart, jsii.String("ollama-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("ollama"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"ollama.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "ollama", "port": 11434, "weight": 1},
				},
			},
		},
	}))

	return chart
}
