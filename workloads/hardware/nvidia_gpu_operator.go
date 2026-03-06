package hardware

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

// NewNvidiaDevicePluginChart deploys the standalone NVIDIA k8s-device-plugin with
// Node Feature Discovery (NFD) and GPU Feature Discovery (GFD) sub-charts.
//
// On Talos, the GPU driver and container toolkit are provided by Talos system extensions
// (nvidia-open-gpu-kernel-modules-production + nvidia-container-toolkit-production), so
// the full GPU operator is unnecessary and its validator is incompatible with Talos paths.
// The standalone device plugin has no validation init containers and works out of the box.
func NewNvidiaDevicePluginChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
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

	// RuntimeClass: tells containerd to use nvidia-container-runtime for the
	// device plugin and GFD pods. Without this, the default runc runtime ignores
	// NVIDIA_VISIBLE_DEVICES=all (already set in the chart template) and never
	// injects libnvidia-ml.so.1 into the container. On Talos, the NVIDIA libs live
	// at /usr/local/glibc/usr/lib/ (squashfs, not bind-mountable via hostPath), so
	// the only way to make NVML accessible is via the nvidia-container-runtime hook.
	// The handler "nvidia" matches the containerd runtime configured by the
	// nvidia-container-toolkit-production Talos extension.
	cdk8s.NewApiObject(chart, jsii.String("nvidia-runtime-class"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("node.k8s.io/v1"),
		Kind:       jsii.String("RuntimeClass"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String("nvidia"),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/handler"), jsii.String("nvidia")))

	cdk8s.NewHelm(chart, jsii.String("nvidia-device-plugin-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("nvidia-device-plugin"),
		Repo:        jsii.String("https://nvidia.github.io/k8s-device-plugin"),
		Version:     jsii.String("0.18.2"),
		ReleaseName: jsii.String("nvidia-device-plugin"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
			// NFD labels GPU nodes with feature.node.kubernetes.io/pci-10de.present=true,
			// which the device plugin DaemonSet uses as its default node affinity.
			"nfd": map[string]any{"enabled": true},
			// GFD adds nvidia.com/gpu.present=true and product/memory/count labels.
			"gfd": map[string]any{"enabled": true},
			// Use the nvidia RuntimeClass so containerd's nvidia-container-runtime hook
			// fires and injects libnvidia-ml.so.1 into the container at startup.
			"runtimeClassName": "nvidia",
			// Inline config avoids a separate ConfigMap resource.
			// - deviceDiscoveryStrategy: nvml — "auto" fails on Talos because it probes
			//   standard paths that don't exist; nvml talks directly to the kernel module
			//   loaded by the nvidia-open-gpu-kernel-modules-production extension.
			// - plugin.deviceListStrategy: envvar — CDI hostPath mounts fail on Talos
			//   (libs at non-standard paths); envvar injects NVIDIA_VISIBLE_DEVICES instead.
			//   NOTE: plugin is nested under flags, not a top-level key (per v1 config schema).
			// - timeSlicing replicas: 2 — Ollama and ComfyUI share the RTX 5070 Ti
			//   (VRAM is not isolated; ~4 GB + ~6 GB fits within the 16 GB pool).
			"config": map[string]any{
				"default": "default",
				"map": map[string]any{
					"default": "version: v1\n" +
						"flags:\n" +
						"  migStrategy: none\n" +
						"  deviceDiscoveryStrategy: nvml\n" +
						"  plugin:\n" +
						"    deviceListStrategy: envvar\n" +
						"sharing:\n" +
						"  timeSlicing:\n" +
						"    resources:\n" +
						"    - name: nvidia.com/gpu\n" +
						"      replicas: 2\n",
				},
			},
		},
	})

	return chart
}
