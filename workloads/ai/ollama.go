package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/imports/ollama"
)

// Base weights for the `executor` model used by the aider offload workflow.
// Kept in sync with FROM in local-ai-setup/ollama/models/executor.Modelfile.
const executorBaseModel = "hf.co/unsloth/Qwen3-Coder-30B-A3B-Instruct-GGUF:UD-Q3_K_XL"

// Storage class for the Ollama model directory. Model blobs are a re-pullable
// cache, so one local replica beats Longhorn's default three-way replication:
// it keeps ~20GB off the other nodes' disks and keeps a 13.8GB model load off
// the iSCSI path.
const modelCacheStorageClass = "longhorn-model-cache"

// NewOllamaChart deploys Ollama LLM inference server via the official Helm chart.
//
// Ollama is pinned to the GPU worker node via nodeSelector nvidia.com/gpu.present=true
// (set by Node Feature Discovery). Time-slicing is configured in the GPU operator
// (5 virtual GPUs), so Ollama shares the RTX 5070 Ti with ComfyUI and Kubeflow workloads.
// An HTTPRoute exposes Ollama at ollama.madhan.app through the homelab Gateway.
func NewOllamaChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// nvidia.com/gpu.present is set by NFD — matches the selector used by the GPU operator.
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
					// memory here is host RAM (cgroup limit), NOT GPU VRAM.
					// GPU VRAM (16GB) is fully available via nvidia.com/gpu: 1.
					// host RAM cgroup limit (worker4 allocatable ~15.1Gi). The executor's
					// 13.8GB GGUF is mmap'd on load and those pages are charged here.
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
			"service": map[string]any{
				"type": "LoadBalancer",
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
