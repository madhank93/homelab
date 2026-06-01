package networking

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

// NewNetbirdPeerChart deploys the NetBird routing peer (k8s-routing-peer) as a
// StatefulSet on any available worker (no nodeSelector — floats for resilience).
//
// The routing peer forms a WireGuard mesh with the Bifrost VPS and advertises the
// 192.168.1.0/24 cluster subnet to Bifrost. Incoming traffic from Bifrost arrives
// on wt0, is forwarded via kernel IP routing, and exits through eth0 on another node
// where Cilium BPF handles the DNAT to the backend pod.
//
// Key configuration decisions:
//   - PVC mounts at /var/lib/netbird/ for peer-key persistence across restarts.
//   - An initContainer applies iptables MASQUERADE on the CILIUM_POST_nat chain
//     so that forwarded traffic from wt0 appears to originate from the node IP.
//   - NB_SKIP_SOCKET_MARK must NOT be set: the routing peer relies on socket fwmark
//     to bypass the management-server host route added to the kernel routing table.
//   - wt0 is excluded from Cilium devices (core/platform/cilium.go) — it is
//     NOARP/POINTOPOINT (no Ethernet header) and would cause Cilium TC BPF to
//     silently drop all traffic without monitor events.
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
	// The setup key is only used on first registration; subsequent restarts use the persisted
	// private key from the netbird-config PVC and ignore the setup key.
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

	// Headless Service — required by StatefulSet spec.serviceName.
	k8s.NewKubeService(chart, jsii.String("netbird-peer-svc"), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("netbird-peer"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.ServiceSpec{
			ClusterIp: jsii.String("None"),
			Selector: &map[string]*string{
				"app": jsii.String("netbird-peer"),
			},
		},
	})

	// StatefulSet: k8s-routing-peer — replaces the old Deployment.
	// Using StatefulSet + PVC so /etc/netbird/ (private key + config) persists across
	// pod restarts. This prevents new peer registrations on every restart, eliminating
	// duplicate peers accumulating in the NetBird Management UI.
	// hostNetwork: true so WireGuard can manipulate the host routing table.
	// Routes are assigned to this peer via the NetBird Management UI (Network → Routes).
	replicas := float64(1)
	k8s.NewKubeStatefulSet(chart, jsii.String("netbird-peer"), &k8s.KubeStatefulSetProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("netbird-peer"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: jsii.String("netbird-peer"),
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
					// Init container: MASQUERADE wt0-forwarded traffic so it exits ens18
					// with this node's LAN IP, enabling Cilium's cil_from_netdev TC BPF
					// on the L2 winner to set the TPROXY mark and redirect to Envoy.
					//
					// Key constraint: kernel WireGuard delivers decapsulated inner packets
					// DIRECTLY to the FORWARD hook, bypassing PREROUTING entirely. So
					// -i wt0 in PREROUTING always sees 0 packets — match by output interface
					// (-o ens18) in POSTROUTING instead. Pod traffic is already handled by
					// CILIUM_POST_nat (runs first) so iptables deduplicates the NAT entry
					// and our rule is a safe no-op for already-masqueraded packets.
					// The -C check prevents duplicate rules on pod restart.
					InitContainers: &[]*k8s.Container{
						{
							Name:    jsii.String("setup-iptables"),
							Image:   jsii.String("netbirdio/netbird:0.71.4"),
							Command: &[]*string{jsii.String("/bin/sh")},
							Args: &[]*string{
								jsii.String("-c"),
								jsii.String("iptables -t nat -C POSTROUTING -o ens18 -d 192.168.1.0/24 -j MASQUERADE 2>/dev/null || iptables -t nat -A POSTROUTING -o ens18 -d 192.168.1.0/24 -j MASQUERADE"),
							},
							SecurityContext: &k8s.SecurityContext{
								Privileged: jsii.Bool(true),
								Capabilities: &k8s.Capabilities{
									Add: &[]*string{jsii.String("NET_ADMIN")},
								},
							},
						},
					},
					Containers: &[]*k8s.Container{
						{
							Name: jsii.String("netbird"),
							// Pinned to match Bifrost server version. Default entrypoint starts
							// the service daemon then calls 'netbird up' — do not override Command.
							Image: jsii.String("netbirdio/netbird:0.71.4"),
							Env: &[]*k8s.EnvVar{
								{
									// Setup key only used on first registration.
									// Read from k8s Secret synced by SecretProviderClass above.
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
								Privileged: jsii.Bool(true),
								Capabilities: &k8s.Capabilities{
									Add: &[]*string{
										jsii.String("NET_ADMIN"),
										jsii.String("SYS_MODULE"),
									},
								},
							},
							VolumeMounts: &[]*k8s.VolumeMount{
								{
									// Persists NetBird private key + config across restarts.
									// Without this, every restart = new key = new peer registration.
									Name:      jsii.String("netbird-config"),
									MountPath: jsii.String("/var/lib/netbird"),
								},
								{
									// CSI mount triggers secretObjects sync → creates netbird-setup-key Secret.
									Name:      jsii.String("openbao-secrets"),
									MountPath: jsii.String("/mnt/secrets"),
									ReadOnly:  jsii.Bool(true),
								},
							},
						},
					},
				},
			},
			// PVC template: 100Mi for /etc/netbird/ config persistence.
			// Longhorn default StorageClass is used automatically.
			VolumeClaimTemplates: &[]*k8s.KubePersistentVolumeClaimProps{
				{
					Metadata: &k8s.ObjectMeta{
						Name: jsii.String("netbird-config"),
					},
					Spec: &k8s.PersistentVolumeClaimSpec{
						AccessModes: &[]*string{jsii.String("ReadWriteOnce")},
						Resources: &k8s.VolumeResourceRequirements{
							Requests: &map[string]k8s.Quantity{
								"storage": k8s.Quantity_FromString(jsii.String("100Mi")),
							},
						},
					},
				},
			},
		},
	})

	return chart
}
