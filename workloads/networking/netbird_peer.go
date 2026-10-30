package networking

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewNetbirdPeerChart deploys the NetBird routing peer (k8s-routing-peer) as a
// StatefulSet on any available worker (no nodeSelector — floats for resilience).
//
// The routing peer forms a WireGuard mesh with the Bifrost VPS and advertises the
// cluster LAN subnet to Bifrost. Incoming traffic from Bifrost arrives
// on wt0, is forwarded via kernel IP routing, and exits the LAN NIC toward another node
// where Cilium BPF handles the DNAT to the backend pod.
//
// Key configuration decisions:
//   - PVC mounts at /var/lib/netbird/ for peer-key persistence across restarts.
//   - An initContainer adds an iptables POSTROUTING MASQUERADE so forwarded traffic
//     from wt0 leaves as the node IP; replies need no route back to the overlay.
//   - NB_SKIP_SOCKET_MARK must NOT be set: the routing peer relies on socket fwmark
//     to bypass the management-server host route added to the kernel routing table.
//   - wt0 is excluded from Cilium devices (core/platform/cilium.go) — it is
//     NOARP/POINTOPOINT (no Ethernet header) and would cause Cilium TC BPF to
//     silently drop all traffic without monitor events.
func NewNetbirdPeerChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Privileged: the peer needs NET_ADMIN and SYS_MODULE for WireGuard.
	kit.Namespace(chart, namespace, true)

	// The setup key is only used on first registration; restarts reuse the
	// private key persisted on the netbird-config PVC.
	kit.SecretProviderClass(chart, "netbird-secrets", "netbird", []string{"NETBIRD_SETUP_KEY"}, "netbird-setup-key",
		kit.Sync{Object: "NETBIRD_SETUP_KEY", Key: "NETBIRD_SETUP_KEY"})

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

	// StatefulSet + PVC keeps the peer's private key across restarts, so a restart
	// does not register a duplicate peer. hostNetwork lets WireGuard manipulate the
	// host routing table. Routes are assigned to this peer in the NetBird UI.
	image := config.ContainerImage("netbird").Ref()
	subnet := config.C.Cluster.Subnet
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
					Volumes:     &[]*k8s.Volume{kit.SecretsVolume("netbird-secrets")},
					// Init container: MASQUERADE wt0-forwarded traffic to the LAN so it leaves
					// with this node's IP and replies return through this peer. Kernel
					// WireGuard hands decapsulated packets straight to FORWARD, skipping
					// PREROUTING, so the rule matches in POSTROUTING on "not back out wt0"
					// rather than a NIC name, which varies with Talos interface naming.
					// The -C check keeps pod restarts from stacking duplicate rules.
					InitContainers: &[]*k8s.Container{
						{
							Name:    jsii.String("setup-iptables"),
							Image:   jsii.String(image),
							Command: &[]*string{jsii.String("/bin/sh")},
							Args: &[]*string{
								jsii.String("-c"),
								jsii.String("iptables -t nat -C POSTROUTING ! -o wt0 -d " + subnet + " -j MASQUERADE 2>/dev/null || iptables -t nat -A POSTROUTING ! -o wt0 -d " + subnet + " -j MASQUERADE"),
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
							Image: jsii.String(image),
							Env: &[]*k8s.EnvVar{
								{
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
									Value: jsii.String("https://" + config.C.Host("netbird")),
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
								// The CSI mount is what creates the netbird-setup-key Secret.
								kit.SecretsMount(),
							},
						},
					},
				},
			},
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
