package compliance

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/k8s"
)

func NewFalcoChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// privileged: Falco DaemonSet pods need host-level access to observe syscalls
	k8s.NewKubeNamespace(chart, jsii.String("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
			Labels: &map[string]*string{
				"pod-security.kubernetes.io/enforce": jsii.String("privileged"),
			},
		},
	})

	values := map[string]any{
		"driver": map[string]any{
			// modern_ebpf: required on Talos Linux — Talos locks down kernel module loading
			// so the kmod and legacy ebpf drivers cannot be used. modern_ebpf uses
			// CO-RE eBPF programs that load without a pre-compiled kernel module.
			"kind": "modern_ebpf",
		},
		"falco": map[string]any{
			"grpc": map[string]any{
				"enabled": true,
			},
			"grpcOutput": map[string]any{
				"enabled": true,
			},
			// json_output: easier to parse and forward to VictoriaLogs via OTel
			"jsonOutput":         true,
			"jsonIncludeOutputProperty": true,
		},
		// falcosidekick: disabled for now — enable to route alerts to Slack/PagerDuty/webhook
		"falcosidekick": map[string]any{
			"enabled": false,
		},
		"resources": map[string]any{
			"limits":   map[string]any{"cpu": "1000m", "memory": "1024Mi"},
			"requests": map[string]any{"cpu": "100m", "memory": "512Mi"},
		},
		// Tolerate all nodes including control plane so every node is protected
		"tolerations": []map[string]any{
			{"operator": "Exists"},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("falco-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("falco"),
		Repo:        jsii.String("https://falcosecurity.github.io/charts"),
		Version:     jsii.String("4.8.0"),
		ReleaseName: jsii.String("falco"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
