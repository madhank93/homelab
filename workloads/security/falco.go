package security

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewFalcoChart deploys Falco runtime security via the official Helm chart.
//
// Falco runs as a DaemonSet and observes Linux system calls to detect anomalous
// container behaviour. The namespace is labelled privileged because the DaemonSet
// pods need host-level access (hostPID, hostNetwork, the modern eBPF probe).
// Falco Sidekick receives alerts over the chart-wired http_output.
func NewFalcoChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, true)

	// The UI login comes from OpenBao, never from chart values: the chart renders
	// webui.user into a Secret, and synth output is published to a public branch.
	kit.SecretProviderClass(chart, "falco-secrets", "falco", []string{"FALCOSIDEKICK_UI_USER"}, "falcosidekick-ui-auth",
		kit.Sync{Object: "FALCOSIDEKICK_UI_USER", Key: "FALCOSIDEKICK_UI_USER"})
	// The falcosidekick UI chart has no extraVolumes to trigger the sync itself.
	kit.SecretSync(chart, "falco-secrets")

	values := map[string]any{
		"podAnnotations": map[string]any{
			"reloader.stakater.com/auto": "true",
		},
		"driver": map[string]any{
			// Talos blocks kernel module loading, so only the CO-RE modern_ebpf
			// probe works; it needs /sys/kernel mounted to find BTF.
			"kind":           "modern_ebpf",
			"sysfsMountPath": "/sys/kernel",
		},
		"falco": map[string]any{
			// JSON output is forwarded to VictoriaLogs via OTel.
			"json_output":                  true,
			"json_include_output_property": true,
		},
		"falcosidekick": map[string]any{
			"enabled": true,
			"webui": map[string]any{
				"enabled": true,
				// Chart key is replicaCount; "replicas" is silently ignored and
				// leaves the chart default of 2.
				"replicaCount": 1,
				// Without a TTL the events fill the fixed 1Gi Redis volume; past
				// half full Redis can no longer save, answers MISCONF, and the UI's
				// wait-redis init container blocks forever.
				"ttl": "7d",
				// Empty user keeps the chart's admin:admin default out of the
				// rendered Secret; the login comes from falcosidekick-ui-auth.
				"user":           "",
				"existingSecret": "falcosidekick-ui-auth",
				"podAnnotations": map[string]any{
					"reloader.stakater.com/auto": "true",
				},
				"resources": map[string]any{
					"limits":   map[string]any{"cpu": "200m", "memory": "128Mi"},
					"requests": map[string]any{"cpu": "50m", "memory": "64Mi"},
				},
			},
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "200m", "memory": "256Mi"},
				"requests": map[string]any{"cpu": "50m", "memory": "128Mi"},
			},
		},
		"resources": map[string]any{
			"limits":   map[string]any{"cpu": "1000m", "memory": "1024Mi"},
			"requests": map[string]any{"cpu": "100m", "memory": "512Mi"},
		},
		// Every node, control planes included, is monitored.
		"tolerations": []map[string]any{
			{"operator": "Exists"},
		},
	}

	falco := config.HelmChart("falco")
	cdk8s.NewHelm(chart, jsii.String("falco-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("falco"),
		Repo:        jsii.String(falco.Repo),
		Version:     jsii.String(falco.Version),
		ReleaseName: jsii.String("falco"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	cdk8s.NewApiObject(chart, jsii.String("falco-servicemonitor"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("monitoring.coreos.com/v1"),
		Kind:       jsii.String("ServiceMonitor"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String("falco-falcosidekick")},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"selector": map[string]any{
			"matchLabels": map[string]any{"app.kubernetes.io/name": "falcosidekick"},
		},
		"namespaceSelector": map[string]any{
			"matchNames": []string{namespace},
		},
		"endpoints": []map[string]any{
			{"port": "http", "path": "/metrics", "interval": "30s"},
		},
	}))

	kit.Route(chart, "falco", "falco", "falco-falcosidekick-ui", 2802)

	return chart
}
