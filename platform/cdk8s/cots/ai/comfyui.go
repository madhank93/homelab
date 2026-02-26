package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/k8s"
)

func NewComfyUIChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
		},
	})

	// PVC for models, outputs, and custom nodes — stored on Longhorn for persistence across restarts
	storageClass := "longhorn"
	k8s.NewKubePersistentVolumeClaim(chart, jsii.String("comfyui-data"), &k8s.KubePersistentVolumeClaimProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("comfyui-data"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.PersistentVolumeClaimSpec{
			AccessModes:      &[]*string{jsii.String("ReadWriteOnce")},
			StorageClassName: &storageClass,
			Resources: &k8s.VolumeResourceRequirements{
				Requests: &map[string]k8s.Quantity{
					"storage": k8s.Quantity_FromString(jsii.String("100Gi")),
				},
			},
		},
	})

	replicas := float64(1)
	k8s.NewKubeDeployment(chart, jsii.String("comfyui"), &k8s.KubeDeploymentProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("comfyui"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.DeploymentSpec{
			Replicas: &replicas,
			// Recreate: GPU workloads cannot have two pods claiming nvidia.com/gpu simultaneously
			Strategy: &k8s.DeploymentStrategy{Type: jsii.String("Recreate")},
			Selector: &k8s.LabelSelector{
				MatchLabels: &map[string]*string{"app": jsii.String("comfyui")},
			},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{
					Labels: &map[string]*string{"app": jsii.String("comfyui")},
				},
				Spec: &k8s.PodSpec{
					// nvidia-container-runtime handles GPU injection via Talos-aware paths
					RuntimeClassName: jsii.String("nvidia"),
					NodeSelector:     &map[string]*string{"nvidia.com/gpu.present": jsii.String("true")},
					Containers: &[]*k8s.Container{
						{
							Name: jsii.String("comfyui"),
							// cu128 matches the CUDA 12.8 userspace provided by driver 570.x
							Image:           jsii.String("yanwk/comfyui-boot:latest-cu128"),
							ImagePullPolicy: jsii.String("IfNotPresent"),
							Ports: &[]*k8s.ContainerPort{
								{ContainerPort: jsii.Number(8188), Name: jsii.String("http"), Protocol: jsii.String("TCP")},
							},
							Env: &[]*k8s.EnvVar{
								{Name: jsii.String("NVIDIA_VISIBLE_DEVICES"), Value: jsii.String("all")},
								// --listen 0.0.0.0: bind on all interfaces so the Service can reach it
								{Name: jsii.String("CLI_ARGS"), Value: jsii.String("--listen 0.0.0.0 --port 8188")},
							},
							Resources: &k8s.ResourceRequirements{
								Limits: &map[string]k8s.Quantity{
									"nvidia.com/gpu": k8s.Quantity_FromNumber(jsii.Number(1)),
									"memory":         k8s.Quantity_FromString(jsii.String("8Gi")),
									"cpu":            k8s.Quantity_FromString(jsii.String("4000m")),
								},
								Requests: &map[string]k8s.Quantity{
									"memory": k8s.Quantity_FromString(jsii.String("4Gi")),
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

	// Service
	cdk8s.NewApiObject(chart, jsii.String("comfyui-service"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Service"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("comfyui"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"selector": map[string]any{"app": "comfyui"},
		"ports": []map[string]any{
			{"name": "http", "port": 8188, "targetPort": 8188, "protocol": "TCP"},
		},
	}))

	// Gateway API HTTPRoute — routes comfyui.madhan.app → comfyui:8188
	cdk8s.NewApiObject(chart, jsii.String("comfyui-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("comfyui"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"comfyui.madhan.app"},
		"rules": []map[string]any{
			{
				"matches":     []map[string]any{{"path": map[string]any{"type": "PathPrefix", "value": "/"}}},
				"backendRefs": []map[string]any{{"name": "comfyui", "port": 8188}},
			},
		},
	}))

	return chart
}
