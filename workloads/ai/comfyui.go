package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewComfyUIChart deploys ComfyUI, a node-based Stable Diffusion UI, on the GPU
// node at comfyui.<domain>. It is scaled to zero by default; `just comfyui on`
// starts it. Models and custom nodes persist on a PVC.
func NewComfyUIChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, false)

	storageClass := "longhorn"
	k8s.NewKubePersistentVolumeClaim(chart, jsii.String("comfyui-data"), &k8s.KubePersistentVolumeClaimProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("comfyui-data"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.PersistentVolumeClaimSpec{
			AccessModes:      &[]*string{jsii.String("ReadWriteMany")},
			StorageClassName: &storageClass,
			Resources: &k8s.VolumeResourceRequirements{
				Requests: &map[string]k8s.Quantity{
					"storage": k8s.Quantity_FromString(jsii.String("50Gi")),
				},
			},
		},
	})

	// Default off: the GPU is shared and unisolated, so the safe resting state
	// leaves its VRAM free. `just comfyui on|off` rewrites this literal.
	replicas := float64(0) // comfyui-replicas
	k8s.NewKubeDeployment(chart, jsii.String("comfyui"), &k8s.KubeDeploymentProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("comfyui"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.DeploymentSpec{
			Replicas: &replicas,
			// Recreate: never run two ComfyUI pods against the shared GPU at once.
			Strategy: &k8s.DeploymentStrategy{Type: jsii.String("Recreate")},
			Selector: &k8s.LabelSelector{
				MatchLabels: &map[string]*string{"app": jsii.String("comfyui")},
			},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{
					Labels: &map[string]*string{"app": jsii.String("comfyui")},
				},
				Spec: &k8s.PodSpec{
					// Without the nvidia runtime the container gets no CUDA libraries.
					RuntimeClassName: jsii.String("nvidia"),
					NodeSelector:     &map[string]*string{"nvidia.com/gpu.present": jsii.String("true")},
					Tolerations: &[]*k8s.Toleration{
						{Key: jsii.String("dedicated"), Operator: jsii.String("Equal"), Value: jsii.String("ai"), Effect: jsii.String("NoSchedule")},
					},
					Containers: &[]*k8s.Container{
						{
							Name: jsii.String("comfyui"),
							// The RTX 5070 Ti is Blackwell (sm_120): the image needs CUDA 12.8 or later.
							Image:           jsii.String(config.ContainerImage("comfyui").Ref()),
							ImagePullPolicy: jsii.String("IfNotPresent"),
							Ports: &[]*k8s.ContainerPort{
								{ContainerPort: jsii.Number(8188), Name: jsii.String("http"), Protocol: jsii.String("TCP")},
							},
							Env: &[]*k8s.EnvVar{
								{Name: jsii.String("NVIDIA_VISIBLE_DEVICES"), Value: jsii.String("all")},
								// ComfyUI binds to localhost unless told otherwise.
								{Name: jsii.String("CLI_ARGS"), Value: jsii.String("--listen 0.0.0.0 --port 8188")},
							},
							Resources: &k8s.ResourceRequirements{
								// 6Gi caps a runaway workflow so it is OOM-killed on its own.
								// With Ollama's 12Gi the limits overcommit worker4's ~15Gi, but
								// most of Ollama's usage is reclaimable mmap'd model pages.
								// The GPU's 16Gi VRAM is not covered by any of this — no
								// Kubernetes limit applies to it.
								Limits: &map[string]k8s.Quantity{
									"nvidia.com/gpu": k8s.Quantity_FromNumber(jsii.Number(1)),
									"memory":         k8s.Quantity_FromString(jsii.String("6Gi")),
									"cpu":            k8s.Quantity_FromString(jsii.String("4000m")),
								},
								Requests: &map[string]k8s.Quantity{
									"memory": k8s.Quantity_FromString(jsii.String("1Gi")),
									"cpu":    k8s.Quantity_FromString(jsii.String("1000m")),
								},
							},
							VolumeMounts: &[]*k8s.VolumeMount{
								{Name: jsii.String("data"), MountPath: jsii.String("/home/user/opt/ComfyUI")},
							},
						},
					},
					Volumes: &[]*k8s.Volume{
						{
							Name: jsii.String("data"),
							PersistentVolumeClaim: &k8s.PersistentVolumeClaimVolumeSource{
								ClaimName: jsii.String("comfyui-data"),
							},
						},
					},
				},
			},
		},
	})

	k8s.NewKubeService(chart, jsii.String("comfyui-service"), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("comfyui"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.ServiceSpec{
			Selector: &map[string]*string{"app": jsii.String("comfyui")},
			Ports: &[]*k8s.ServicePort{
				{
					Name:       jsii.String("http"),
					Port:       jsii.Number(8188),
					TargetPort: k8s.IntOrString_FromNumber(jsii.Number(8188)),
					Protocol:   jsii.String("TCP"),
				},
			},
		},
	})

	kit.Route(chart, "comfyui", "comfyui", "comfyui", 8188)

	return chart
}
