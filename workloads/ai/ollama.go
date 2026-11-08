// Package ai deploys the GPU-backed AI workloads on worker4: Ollama, Open
// WebUI, ComfyUI and Kubeflow.
package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/imports/ollama"
	"github.com/madhank93/homelab/workloads/kit"
)

// executorBaseModel is the base of the `executor` model used by the aider
// offload workflow; it must match FROM in local-ai-setup/ollama/models/executor.Modelfile.
const executorBaseModel = "hf.co/unsloth/Qwen3-Coder-30B-A3B-Instruct-GGUF:UD-Q3_K_XL"

// modelCacheStorageClass backs the Ollama model directory. Model blobs are a re-pullable
// cache, so one local replica beats Longhorn's default three-way replication:
// it keeps ~20GB off the other nodes' disks and keeps a 13.8GB model load off
// the iSCSI path.
const modelCacheStorageClass = "longhorn-model-cache"

// NewOllamaChart deploys the Ollama inference server via its Helm chart, pinned
// to the GPU node. The device plugin time-slices the GPU, so Ollama shares it
// with ComfyUI and Kubeflow. It is reachable at ollama.<domain> and on a fixed
// LoadBalancer IP.
func NewOllamaChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// nvidia.com/gpu.present is set by GPU Feature Discovery (hardware package).
	gpuNodeSelector := map[string]any{
		"nvidia.com/gpu.present": "true",
	}

	k8s.NewKubeStorageClass(chart, jsii.String("model-cache-storageclass"), &k8s.KubeStorageClassProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(modelCacheStorageClass),
		},
		Provisioner:          jsii.String("driver.longhorn.io"),
		ReclaimPolicy:        jsii.String("Delete"),
		AllowVolumeExpansion: jsii.Bool(true),
		VolumeBindingMode:    jsii.String("Immediate"),
		Parameters: &map[string]*string{
			"numberOfReplicas": jsii.String("1"),
			// strict-local pins the single replica to the pod's node, so the GPU
			// node reads model files from its own disk.
			"dataLocality":        jsii.String("strict-local"),
			"staleReplicaTimeout": jsii.String("30"),
			"fsType":              jsii.String("ext4"),
		},
	})

	ollama.NewOllama(chart, jsii.String("ollama-release"), &ollama.OllamaProps{
		ReleaseName: jsii.String("ollama"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
			// Image tag intentionally unset: the chart default tracks its appVersion,
			// so bumping the chart carries the Ollama runtime with it.
			"replicaCount": 1,
			"resources": map[string]any{
				"limits": map[string]any{
					"nvidia.com/gpu": 1,
					// Host RAM, not VRAM (worker4 allocatable ~15.1Gi). The executor's
					// 13.8GB GGUF is mmap'd on load and those pages count against it.
					"memory": "12Gi",
					"cpu":    "4000m",
				},
				"requests": map[string]any{
					"memory": "4Gi",
					"cpu":    "1000m",
				},
			},
			// The chart's key is persistentVolume. An unrecognised key is silently
			// dropped, and the model directory then falls back to the chart's default
			// emptyDir — which wipes every model on pod restart.
			"persistentVolume": map[string]any{
				"enabled":      true,
				"size":         "40Gi",
				"storageClass": modelCacheStorageClass,
				"accessModes":  []string{"ReadWriteOnce"},
			},
			// Fixed IP: the laptop's OLLAMA_API_BASE (aider/offload) points at it.
			"service": map[string]any{
				"type":        "LoadBalancer",
				"port":        11434,
				"annotations": map[string]any{"lbipam.cilium.io/ips": config.C.Cluster.LBIPs.Ollama},
			},
			"nodeSelector": gpuNodeSelector,
			"tolerations": []map[string]any{
				{"key": "dedicated", "operator": "Equal", "value": "ai", "effect": "NoSchedule"},
			},
			// The nvidia runtime comes from the nvidia-container-toolkit-production Talos extension.
			"runtimeClassName": "nvidia",
			"extraEnv": []map[string]any{
				{"name": "NVIDIA_VISIBLE_DEVICES", "value": "all"},
				// The KV cache is the only part of the footprint that grows with
				// context, and on a 16GB card it is what decides how large a
				// prompt the executor can take. q8_0 halves it; flash attention
				// is what makes a quantised KV cache usable at all.
				{"name": "OLLAMA_FLASH_ATTENTION", "value": "1"},
				{"name": "OLLAMA_KV_CACHE_TYPE", "value": "q8_0"},
			},
			// Declarative pull of the executor's base weights into the PVC. The
			// `executor` model itself is an `ollama create` layered on top of this
			// base (Modelfile in the local-ai-setup repo); Helm values can pull a
			// model but cannot build one.
			"ollama": map[string]any{
				"models": map[string]any{
					"pull": []string{executorBaseModel},
				},
			},
		},
	})

	kit.Route(chart, "ollama", "ollama", "ollama", 11434)

	return chart
}
