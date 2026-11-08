// Package hardware deploys GPU support for worker4: the NVIDIA device plugin
// with node feature discovery, the nvidia RuntimeClass and the DCGM exporter.
package hardware

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/kit"
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

	kit.Namespace(chart, namespace, true)

	// Every GPU pod needs this RuntimeClass. Talos keeps the NVIDIA libs on a
	// squashfs that hostPath cannot mount, so only the nvidia-container-runtime
	// hook can inject them; plain runc ignores NVIDIA_VISIBLE_DEVICES. The
	// "nvidia" handler is registered by the nvidia-container-toolkit-production extension.
	k8s.NewKubeRuntimeClass(chart, jsii.String("nvidia-runtime-class"), &k8s.KubeRuntimeClassProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("nvidia")},
		Handler:  jsii.String("nvidia"),
	})

	plugin := config.HelmChart("nvidia-device-plugin")
	cdk8s.NewHelm(chart, jsii.String("nvidia-device-plugin-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("nvidia-device-plugin"),
		Repo:        jsii.String(plugin.Repo),
		Version:     jsii.String(plugin.Version),
		ReleaseName: jsii.String("nvidia-device-plugin"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
			// NFD labels GPU nodes with feature.node.kubernetes.io/pci-10de.present=true,
			// which the device plugin DaemonSet uses as its default node affinity.
			"nfd": map[string]any{
				"enabled": true,
				// Replaces the chart's list rather than merging, so its two defaults are
				// repeated alongside the GPU worker's taint.
				"worker": map[string]any{
					"tolerations": []map[string]any{
						{"key": "node-role.kubernetes.io/master", "operator": "Equal", "value": "", "effect": "NoSchedule"},
						{"key": "nvidia.com/gpu", "operator": "Equal", "value": "present", "effect": "NoSchedule"},
						{"key": "dedicated", "operator": "Equal", "value": "ai", "effect": "NoSchedule"},
					},
				},
			},
			"tolerations": []map[string]any{
				{"key": "dedicated", "operator": "Equal", "value": "ai", "effect": "NoSchedule"},
			},
			// GFD adds nvidia.com/gpu.present=true and product/memory/count labels.
			"gfd":              map[string]any{"enabled": true},
			"runtimeClassName": "nvidia",
			// - deviceDiscoveryStrategy nvml: "auto" probes standard paths that Talos lacks.
			// - deviceListStrategy envvar: CDI hostPath mounts fail on Talos's library
			//   paths. It sits under flags.plugin in the v1 config schema.
			// - timeSlicing 5: Ollama, ComfyUI, a notebook, a training job and a Katib
			//   trial. VRAM is not isolated, so two large models at once can OOM the card.
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
						"      replicas: 5\n",
				},
			},
		},
	})

	// DCGM exporter: GPU utilisation, VRAM, temperature and power, scraped through
	// its ServiceMonitor. Needs the nvidia runtime to load NVML.
	dcgm := config.HelmChart("dcgm-exporter")
	cdk8s.NewHelm(chart, jsii.String("dcgm-exporter-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("dcgm-exporter"),
		Repo:        jsii.String(dcgm.Repo),
		Version:     jsii.String(dcgm.Version),
		ReleaseName: jsii.String("dcgm-exporter"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
			// GFD sets nvidia.com/gpu.present on GPU nodes.
			"affinity": map[string]any{
				"nodeAffinity": map[string]any{
					"requiredDuringSchedulingIgnoredDuringExecution": map[string]any{
						"nodeSelectorTerms": []map[string]any{
							{
								"matchExpressions": []map[string]any{
									{"key": "nvidia.com/gpu.present", "operator": "In", "values": []string{"true"}},
								},
							},
						},
					},
				},
			},
			"tolerations": []map[string]any{
				{"key": "dedicated", "operator": "Equal", "value": "ai", "effect": "NoSchedule"},
			},
			"runtimeClassName": "nvidia",
			"serviceMonitor": map[string]any{
				"enabled": true,
			},
		},
	})

	return chart
}
