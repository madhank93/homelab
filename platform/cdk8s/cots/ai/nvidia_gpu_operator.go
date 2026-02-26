package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/k8s"
)

func NewNvidiaGpuOperatorChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
			Labels: &map[string]*string{
				"pod-security.kubernetes.io/enforce": jsii.String("privileged"),
			},
		},
	})

	// nvidia.com/gpu.present is set automatically by NFD on nodes with a detected GPU.
	// This is more reliable than a manually applied node-role label.
	nodeSelector := map[string]any{
		"nvidia.com/gpu.present": jsii.String("true"),
	}

	values := map[string]any{
		// driver.enabled=true: required so the GPU operator can manage driver lifecycle.
		// In practice, because the Talos nvidia-open-gpu-kernel-modules-production extension
		// is present (detected via extensions.talos.dev/* node labels), the GPU operator
		// automatically labels the node nvidia.com/gpu.deploy.driver=pre-installed and sets
		// the driver DaemonSet DESIRED=0. The driver container never runs; the Talos extension
		// provides both kernel modules and userspace CUDA libs at /usr/local/glibc/usr/lib/.
		"driver": map[string]any{
			"enabled":      true,
			"nodeSelector": nodeSelector,
		},
		// toolkit.enabled=false: the Talos nvidia-container-toolkit-production extension
		// installs toolkit binaries at /usr/local/bin/ and writes
		// /etc/cri/conf.d/10-nvidia-container-runtime.part automatically.
		// The GPU operator toolkit DaemonSet is redundant on Talos.
		"toolkit": map[string]any{
			"enabled":      false,
			"nodeSelector": nodeSelector,
		},
		"operator": map[string]any{
			"defaultRuntime": "containerd",
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
		},
		// DEVICE_LIST_STRATEGY=envvars: use env-var injection instead of CDI.
		// CDI mode generates spec entries with hostPath=/usr/lib/libX.so which doesn't
		// exist on Talos (libs live at /usr/local/glibc/usr/lib/). With envvars mode the
		// device plugin injects NVIDIA_VISIBLE_DEVICES=<uuid> into the container, and the
		// Talos nvidia-container-runtime (from nvidia-container-toolkit-production extension)
		// handles GPU device injection using its own Talos-aware library paths.
		"devicePlugin": map[string]any{
			"nodeSelector": nodeSelector,
			"env": []map[string]any{
				{"name": "DEVICE_LIST_STRATEGY", "value": "envvars"},
				{"name": "CDI_ENABLED", "value": "false"},
			},
		},
		"dcgmExporter": map[string]any{"nodeSelector": nodeSelector},
		"gfd":          map[string]any{"nodeSelector": nodeSelector},
	}

	// Talos Validation Bridge DaemonSet
	// GPU operator v25.x uses a validation chain (driver-ready, toolkit-ready) to coordinate
	// startup between its components. On Talos, the Talos extensions (nvidia-open-gpu-kernel-
	// modules-production, nvidia-container-toolkit-production) provide driver and toolkit without
	// the GPU operator's own driver/toolkit containers. This DaemonSet creates the three marker
	// files that GPU operator components watch to confirm readiness:
	//   .driver-ctr-ready  — signals driver-validation init container to proceed
	//   driver-ready       — signals driver is available
	//   toolkit-ready      — signals device-plugin toolkit-validation init container to proceed
	// Without these files, the device plugin stays in Init:0/1 forever.
	trueBool := true
	privileged := true
	zero := float64(0)
	k8s.NewKubeDaemonSet(chart, jsii.String("talos-validation-bridge"), &k8s.KubeDaemonSetProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("talos-nvidia-validation-bridge"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.DaemonSetSpec{
			Selector: &k8s.LabelSelector{
				MatchLabels: &map[string]*string{
					"app": jsii.String("talos-nvidia-validation-bridge"),
				},
			},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{
					Labels: &map[string]*string{
						"app": jsii.String("talos-nvidia-validation-bridge"),
					},
				},
				Spec: &k8s.PodSpec{
					NodeSelector:                  &map[string]*string{"nvidia.com/gpu.present": jsii.String("true")},
					AutomountServiceAccountToken:  &trueBool,
					TerminationGracePeriodSeconds: &zero,
					Containers: &[]*k8s.Container{
						{
							Name:  jsii.String("bridge"),
							Image: jsii.String("busybox:1.37"),
							Command: &[]*string{
								jsii.String("/bin/sh"), jsii.String("-c"),
							},
							Args: &[]*string{jsii.String(
								"mkdir -p /run/nvidia/validations && " +
									"touch /run/nvidia/validations/.driver-ctr-ready " +
									"/run/nvidia/validations/driver-ready " +
									"/run/nvidia/validations/toolkit-ready && " +
									"while true; do sleep 3600; done",
							)},
							SecurityContext: &k8s.SecurityContext{
								Privileged: &privileged,
							},
							VolumeMounts: &[]*k8s.VolumeMount{
								{
									Name:      jsii.String("run-nvidia-validations"),
									MountPath: jsii.String("/run/nvidia/validations"),
								},
							},
						},
					},
					Volumes: &[]*k8s.Volume{
						{
							Name: jsii.String("run-nvidia-validations"),
							HostPath: &k8s.HostPathVolumeSource{
								Path: jsii.String("/run/nvidia/validations"),
								Type: jsii.String("DirectoryOrCreate"),
							},
						},
					},
				},
			},
		},
	})

	// CRDs
	cdk8s.NewInclude(chart, jsii.String("nvidia-cluster-policy-crd"), &cdk8s.IncludeProps{
		Url: jsii.String("https://raw.githubusercontent.com/NVIDIA/gpu-operator/v25.10.1/deployments/gpu-operator/crds/nvidia.com_clusterpolicies.yaml"),
	})
	cdk8s.NewInclude(chart, jsii.String("nvidia-driver-crd"), &cdk8s.IncludeProps{
		Url: jsii.String("https://raw.githubusercontent.com/NVIDIA/gpu-operator/v25.10.1/deployments/gpu-operator/crds/nvidia.com_nvidiadrivers.yaml"),
	})

	cdk8s.NewHelm(chart, jsii.String("gpu-operator-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("gpu-operator"),
		Repo:        jsii.String("https://helm.ngc.nvidia.com/nvidia"),
		Version:     jsii.String("v25.10.1"),
		ReleaseName: jsii.String("gpu-operator"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
