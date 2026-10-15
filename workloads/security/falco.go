package security

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

// NewFalcoChart deploys Falco runtime security via the official Helm chart.
//
// Falco runs as a DaemonSet and observes Linux system calls to detect anomalous
// container behaviour. The namespace is labelled privileged because the DaemonSet
// pods need host-level access (hostPID, hostNetwork, /dev/falco, kernel module or eBPF probe).
// Falco Sidekick is also deployed to forward alerts to VictoriaLogs.
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

	// The UI login comes from OpenBao, never from chart values: the chart renders
	// webui.user into a Secret, and synth output is published to a public branch.
	cdk8s.NewApiObject(chart, jsii.String("falco-spc"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets-store.csi.x-k8s.io/v1"),
		Kind:       jsii.String("SecretProviderClass"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("falco-secrets"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"provider": "openbao",
		"parameters": map[string]any{
			"vaultAddress": "http://openbao.openbao.svc.cluster.local:8200",
			"roleName":     "falco",
			"objects": `- objectName: "FALCOSIDEKICK_UI_USER"
  secretPath: "secret/data/falco"
  secretKey: "FALCOSIDEKICK_UI_USER"`,
		},
		"secretObjects": []map[string]any{
			{
				"secretName": "falcosidekick-ui-auth",
				"type":       "Opaque",
				"data": []map[string]any{
					{"objectName": "FALCOSIDEKICK_UI_USER", "key": "FALCOSIDEKICK_UI_USER"},
				},
			},
		},
	}))

	k8s.NewKubeServiceAccount(chart, jsii.String("falco-secret-sync-sa"), &k8s.KubeServiceAccountProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("secret-sync"),
			Namespace: jsii.String(namespace),
		},
	})

	// The falcosidekick UI chart has no extraVolumes, so this pod mounts the CSI
	// volume to trigger the secretObjects sync.
	replicas := float64(1)
	k8s.NewKubeDeployment(chart, jsii.String("falco-secret-sync"), &k8s.KubeDeploymentProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("secret-sync"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.DeploymentSpec{
			Replicas: &replicas,
			Selector: &k8s.LabelSelector{
				MatchLabels: &map[string]*string{"app": jsii.String("secret-sync")},
			},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{
					Labels:      &map[string]*string{"app": jsii.String("secret-sync")},
					Annotations: &map[string]*string{"reloader.stakater.com/auto": jsii.String("true")},
				},
				Spec: &k8s.PodSpec{
					ServiceAccountName: jsii.String("secret-sync"),
					Volumes: &[]*k8s.Volume{
						{
							Name: jsii.String("openbao-secrets"),
							Csi: &k8s.CsiVolumeSource{
								Driver:   jsii.String("secrets-store.csi.k8s.io"),
								ReadOnly: jsii.Bool(true),
								VolumeAttributes: &map[string]*string{
									"secretProviderClass": jsii.String("falco-secrets"),
								},
							},
						},
					},
					Containers: &[]*k8s.Container{
						{
							Name:  jsii.String("pause"),
							Image: jsii.String("registry.k8s.io/pause:3.10"),
							VolumeMounts: &[]*k8s.VolumeMount{
								{
									Name:      jsii.String("openbao-secrets"),
									MountPath: jsii.String("/mnt/secrets"),
									ReadOnly:  jsii.Bool(true),
								},
							},
						},
					},
				},
			},
		},
	})

	values := map[string]any{
		"podAnnotations": map[string]any{
			"reloader.stakater.com/auto": "true",
		},
		"driver": map[string]any{
			// modern_ebpf: required on Talos Linux — Talos locks down kernel module loading
			// so the kmod and legacy ebpf drivers cannot be used. modern_ebpf uses
			// CO-RE eBPF programs that load without a pre-compiled kernel module.
			"kind": "modern_ebpf",
			// sysfsMountPath: mounts /sys/kernel into the container so the modern_ebpf
			// probe can find BTF at /sys/kernel/btf/vmlinux. Required on Talos because
			// the container's default sysfs does not expose /sys/kernel/btf.
			"sysfsMountPath": "/sys/kernel",
		},
		"falco": map[string]any{
			"grpc": map[string]any{
				"enabled": true,
			},
			"grpc_output": map[string]any{
				"enabled": true,
			},
			// json_output: easier to parse and forward to VictoriaLogs via OTel
			"json_output":                    true,
			"json_include_output_property":   true,
		},
		"falcosidekick": map[string]any{
			"enabled": true,
			"webui": map[string]any{
				"enabled": true,
				// Chart key is replicaCount; "replicas" is silently ignored and
				// leaves the chart default of 2.
				"replicaCount": 1,
				// Without a TTL the UI keeps every event forever on a fixed 1Gi
				// volume. Redis writes a full temp copy before renaming it over
				// dump.rdb, so once the dataset passes half the volume no save can
				// ever complete: it fails with "No space left on device", leaves the
				// partial temp file behind, and answers MISCONF instead of PONG —
				// which blocks the UI's wait-redis init container permanently.
				"ttl": "7d",
				// Empty user keeps the chart's admin:admin default out of the
				// rendered Secret; the login comes from falcosidekick-ui-auth.
				"user":           "",
				"existingSecret": "falcosidekick-ui-auth",
				// The UI restarts when the synced login changes.
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
		// Tolerate all nodes including control plane so every node is protected
		"tolerations": []map[string]any{
			{"operator": "Exists"},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("falco-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("falco"),
		Repo:        jsii.String("https://falcosecurity.github.io/charts"),
		Version:     jsii.String("8.0.5"),
		ReleaseName: jsii.String("falco"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// ServiceMonitor — VMAgent scrapes falcosidekick metrics on port 2801.
	// Falcosidekick exposes Prometheus metrics at :2801/metrics about forwarded events.
	cdk8s.NewApiObject(chart, jsii.String("falco-servicemonitor"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("monitoring.coreos.com/v1"),
		Kind:       jsii.String("ServiceMonitor"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("falco-falcosidekick"),
			Namespace: jsii.String(namespace),
		},
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

	// Gateway API HTTPRoute — falco.madhan.app → falcosidekick-ui (port 2802)
	cdk8s.NewApiObject(chart, jsii.String("falco-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("falco"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"falco.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "falco-falcosidekick-ui", "port": 2802, "weight": 1},
				},
			},
		},
	}))

	return chart
}
