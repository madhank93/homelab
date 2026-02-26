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
		// driver.enabled=false: Talos loads NVIDIA kernel modules (570.x) via the
		// nvidia-open-gpu-kernel-modules extension. No driver container needed.
		"driver": map[string]any{"enabled": false},
		// toolkit.enabled=false: the Talos nvidia-container-toolkit-production extension
		// installs toolkit binaries at /usr/local/bin/ and writes
		// /etc/cri/conf.d/10-nvidia-container-runtime.part automatically.
		// The GPU operator toolkit DaemonSet is redundant on Talos.
		"toolkit": map[string]any{
			"enabled":      false,
			"nodeSelector": nodeSelector,
		},
		// validator: tell the driver-validation init container that the driver is
		// pre-installed (Talos kernel module), not in a GPU operator driver container.
		// Without DISABLE_DEV_CHAR_SYMLINK_CREATION=true the validator waits forever
		// for /run/nvidia/driver which Talos never populates.
		"validator": map[string]any{
			"driver": map[string]any{
				"env": []map[string]any{
					{"name": "DISABLE_DEV_CHAR_SYMLINK_CREATION", "value": "true"},
				},
			},
		},
		"operator": map[string]any{
			"defaultRuntime": "containerd",
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
		},
		"devicePlugin": map[string]any{"nodeSelector": nodeSelector},
		"dcgmExporter": map[string]any{"nodeSelector": nodeSelector},
		"gfd":          map[string]any{"nodeSelector": nodeSelector},
	}

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
