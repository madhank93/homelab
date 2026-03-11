package networking

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

func NewNetbirdPeerChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Privileged namespace — netbird peer needs NET_ADMIN and SYS_MODULE for WireGuard.
	k8s.NewKubeNamespace(chart, jsii.String("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
			Labels: &map[string]*string{
				"pod-security.kubernetes.io/enforce": jsii.String("privileged"),
			},
		},
	})

	// SecretProviderClass — Pattern B (secretObjects sync).
	// Fetches NETBIRD_SETUP_KEY from OpenBao and syncs it into k8s Secret "netbird-setup-key".
	// The deployment's env var references this secret via secretKeyRef.
	// Prerequisite: write NETBIRD_SETUP_KEY at path secret/data/netbird in OpenBao.
	cdk8s.NewApiObject(chart, jsii.String("netbird-spc"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets-store.csi.x-k8s.io/v1"),
		Kind:       jsii.String("SecretProviderClass"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("netbird-secrets"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"provider": "openbao",
		"parameters": map[string]any{
			"vaultAddress": "http://openbao.openbao.svc.cluster.local:8200",
			"roleName":     "netbird",
			"objects": `- objectName: "NETBIRD_SETUP_KEY"
  secretPath: "secret/data/netbird"
  secretKey: "NETBIRD_SETUP_KEY"`,
		},
		"secretObjects": []map[string]any{
			{
				"secretName": "netbird-setup-key",
				"type":       "Opaque",
				"data": []map[string]any{
					{
						"objectName": "NETBIRD_SETUP_KEY",
						"key":        "NETBIRD_SETUP_KEY",
					},
				},
			},
		},
	}))

	// Deployment: k8s-routing-peer — connects to NetBird mesh and advertises 192.168.1.0/24.
	// hostNetwork: true so WireGuard can manipulate host routing table.
	// dnsPolicy: ClusterFirstWithHostNet to retain in-cluster DNS resolution.
	// Routes are assigned to this peer via the NetBird Management UI (Network → Routes)
	// and pushed down automatically — no CLI flag needed.
	replicas := float64(1)
	k8s.NewKubeDeployment(chart, jsii.String("netbird-peer"), &k8s.KubeDeploymentProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("netbird-peer"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.DeploymentSpec{
			Replicas: &replicas,
			Selector: &k8s.LabelSelector{
				MatchLabels: &map[string]*string{"app": jsii.String("netbird-peer")},
			},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{
					Labels:      &map[string]*string{"app": jsii.String("netbird-peer")},
					Annotations: &map[string]*string{"reloader.stakater.com/auto": jsii.String("true")},
				},
				Spec: &k8s.PodSpec{
					HostNetwork: jsii.Bool(true),
					DnsPolicy:   jsii.String("ClusterFirstWithHostNet"),
					Volumes: &[]*k8s.Volume{
						{
							Name: jsii.String("openbao-secrets"),
							Csi: &k8s.CsiVolumeSource{
								Driver:   jsii.String("secrets-store.csi.k8s.io"),
								ReadOnly: jsii.Bool(true),
								VolumeAttributes: &map[string]*string{
									"secretProviderClass": jsii.String("netbird-secrets"),
								},
							},
						},
					},
					Containers: &[]*k8s.Container{
						{
							Name: jsii.String("netbird"),
							// Pinned to match Bifrost server version. The default entrypoint
							// starts the service daemon then calls 'netbird up' — do not
							// override Command or the daemon mode breaks.
							Image: jsii.String("netbirdio/netbird:0.66.2"),
							Env: &[]*k8s.EnvVar{
								{
									// NB_SETUP_KEY read from the k8s Secret synced by SecretProviderClass.
									// The CSI volume mount below is required to trigger the sync.
									Name: jsii.String("NB_SETUP_KEY"),
									ValueFrom: &k8s.EnvVarSource{
										SecretKeyRef: &k8s.SecretKeySelector{
											Name: jsii.String("netbird-setup-key"),
											Key:  jsii.String("NETBIRD_SETUP_KEY"),
										},
									},
								},
								{
									Name:  jsii.String("NB_MANAGEMENT_URL"),
									Value: jsii.String("https://netbird.madhan.app"),
								},
								{
									Name:  jsii.String("NB_HOSTNAME"),
									Value: jsii.String("k8s-routing-peer"),
								},
							},
							SecurityContext: &k8s.SecurityContext{
								// privileged: required on Talos — capability grants (NET_ADMIN, SYS_MODULE)
								// are blocked by the node security profile without explicit privileged mode.
								Privileged: jsii.Bool(true),
								Capabilities: &k8s.Capabilities{
									Add: &[]*string{
										jsii.String("NET_ADMIN"),
										jsii.String("SYS_MODULE"),
									},
								},
							},
							// CSI volume mount triggers secretObjects sync → creates netbird-setup-key Secret.
							// The mount itself is not used by the container — only the sync matters.
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

	return chart
}
