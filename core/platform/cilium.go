package platform

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// InstallCilium installs Cilium as the cluster CNI and kube-proxy replacement.
//
// Configuration highlights:
//   - kubeProxyReplacement: true — full eBPF data plane; kube-proxy is disabled in Talos
//   - k8sServiceHost/Port: VIP 192.168.1.210:6443 — Cilium contacts API server directly
//   - devices: [ens18, eth0] — covers Talos ≤v1.12 (ens18) and v1.13+ (eth0); wt0 excluded
//     (NOARP/POINTOPOINT — TC BPF silently drops non-Ethernet frames)
//   - l2Announcements + gatewayAPI: bare-metal LoadBalancer IPs and Gateway API support
//   - Hubble relay + UI: network flow observability
//
// An HTTPRoute for hubble.madhan.app → hubble-ui:80 is created after the chart.
func InstallCilium(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {
	ciliumChart, err := helm.NewRelease(ctx, "cilium", &helm.ReleaseArgs{
		Chart:   pulumi.String("cilium"),
		Version: pulumi.String("1.18.10"), // Pinned: 1.19.x regression blocks host TCP on eth0 nodes (cilium/cilium#44430)
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://helm.cilium.io/"),
		},
		Namespace:       pulumi.String("kube-system"),
		CreateNamespace: pulumi.Bool(true),
		Values: pulumi.Map{
			"ipam": pulumi.Map{
				"mode": pulumi.String("kubernetes"),
			},
			"kubeProxyReplacement": pulumi.Bool(true), // Strict eBPF
			"securityContext": pulumi.Map{
				"capabilities": pulumi.Map{
					"ciliumAgent": pulumi.StringArray{
						pulumi.String("CHOWN"), pulumi.String("KILL"), pulumi.String("NET_ADMIN"),
						pulumi.String("NET_RAW"), pulumi.String("IPC_LOCK"), pulumi.String("SYS_ADMIN"),
						pulumi.String("SYS_RESOURCE"), pulumi.String("DAC_OVERRIDE"), pulumi.String("FOWNER"),
						pulumi.String("SETGID"), pulumi.String("SETUID"),
					},
					"cleanCiliumState": pulumi.StringArray{
						pulumi.String("NET_ADMIN"), pulumi.String("SYS_ADMIN"), pulumi.String("SYS_RESOURCE"),
					},
				},
			},
			"cgroup": pulumi.Map{
				"autoMount": pulumi.Map{
					"enabled": pulumi.Bool(false),
				},
				"hostRoot": pulumi.String("/sys/fs/cgroup"),
			},
			"k8sServiceHost": pulumi.String("192.168.1.210"), // VIP
			"k8sServicePort": pulumi.Int(6443),
			"hubble": pulumi.Map{
				"enabled": pulumi.Bool(true),
				"relay": pulumi.Map{
					"enabled": pulumi.Bool(true),
				},
				"ui": pulumi.Map{
					"enabled": pulumi.Bool(true),
				},
			},
			// Access & Routing
			"l2Announcements": pulumi.Map{
				"enabled": pulumi.Bool(true),
			},
			"externalIPs": pulumi.Map{
				"enabled": pulumi.Bool(true),
			},
			"extraConfig": pulumi.Map{
				"enable-l2-announcements": pulumi.String("true"),
				"enable-external-ips":     pulumi.String("true"),
			},
			"gatewayAPI": pulumi.Map{
				"enabled": pulumi.Bool(true),
			},
			// Both ens18 (existing VMs, Talos ≤v1.12 predictable naming) and eth0
			// (fresh VMs from Talos v1.13+ nocloud image, classic naming) listed so
			// Cilium works across both. Cilium skips non-existent interfaces gracefully
			// and auto-detects the direct routing device from whichever one has the
			// node IP. wt0 (NetBird WireGuard) excluded — NOARP/POINTOPOINT; TC BPF
			// silently drops non-Ethernet frames.
			"devices": pulumi.StringArray{
				pulumi.String("ens18"),
				pulumi.String("eth0"),
			},
			// Talos enables forwardKubeDNSToHost by default; Cilium's eBPF host-routing
			// conflicts with this — host-namespace DNS queries get misrouted. Legacy
			// routing uses the kernel routing table for host-namespace traffic instead,
			// which is also required for kubelet (10250) and Talos apid (50000) to be
			// reachable: without it, TC ingress BPF routes packets via cilium_host before
			// the BPF forwarding path is fully initialised, dropping them.
			"bpf": pulumi.Map{
				"hostLegacyRouting": pulumi.Bool(true),
			},
			// Cilium's socket-level LB hooks (cil_sock4_post_bind) intercept bind()
			// calls and interfere with kubelet and Talos apid binding to their ports
			// (10250, 50000) after Cilium starts. Disabling prevents this.
			"hostServices": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
			// Cilium's NOTRACK iptables rules bypass conntrack for all traffic.
			// Combined with DROP-INVALID rules this blocks all NEW TCP to host ports
			// (10250, 50000) while allowing ESTABLISHED. Disabling lets the kernel
			// track host connections normally.
			"installNoConntrackIptablesRules": pulumi.Bool(false),
		},
	}, pulumi.Provider(k8sProvider))
	if err != nil {
		return err
	}

	// HTTPRoute for Hubble UI — routes hubble.madhan.app → hubble-ui:80 in kube-system
	_, err = apiextensions.NewCustomResource(ctx, "hubble-ui-httproute", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("gateway.networking.k8s.io/v1"),
		Kind:       pulumi.String("HTTPRoute"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("hubble-ui"),
			Namespace: pulumi.String("kube-system"),
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"parentRefs": []map[string]any{
					{"name": "homelab-gateway", "namespace": "kube-system"},
				},
				"hostnames": []string{"hubble.madhan.app"},
				"rules": []map[string]any{
					{
						"matches":     []map[string]any{{"path": map[string]any{"type": "PathPrefix", "value": "/"}}},
						"backendRefs": []map[string]any{{"name": "hubble-ui", "port": 80}},
					},
				},
			},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{ciliumChart}))

	return err
}

// InstallGateway creates the shared Gateway resource for the cluster
func InstallGateway(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {

	// Install Gateway API CRDs (Experimental v1.2.1 - Required for Cilium 1.16+)
	crds, err := yaml.NewConfigFile(ctx, "gateway-api-crds", &yaml.ConfigFileArgs{
		File: "https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/experimental-install.yaml",
	}, pulumi.Provider(k8sProvider))
	if err != nil {
		return err
	}

	// Create Gateway Resource (Depends on GatewayClass from Cilium Helm Chart + CRDs)
	_, err = apiextensions.NewCustomResource(ctx, "cilium-gateway", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("gateway.networking.k8s.io/v1"),
		Kind:       pulumi.String("Gateway"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("homelab-gateway"),
			Namespace: pulumi.String("kube-system"),
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"gatewayClassName": "cilium",
				"listeners": []map[string]any{
					{
						"name":     "http",
						"protocol": "HTTP",
						"port":     80,
						"allowedRoutes": map[string]any{
							"namespaces": map[string]any{
								"from": "All",
							},
						},
					},
					{
						"name":     "https",
						"protocol": "HTTPS",
						"port":     443,
						"allowedRoutes": map[string]any{
							"namespaces": map[string]any{
								"from": "All",
							},
						},
						"tls": map[string]any{
							"mode": "Terminate",
							"certificateRefs": []map[string]any{
								{
									"name":      "wildcard-madhan-app-tls",
									"namespace": "kube-system",
								},
							},
						},
					},
				},
			},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{crds}))

	return err
}

// ConfigureCiliumIPPool creates the L2 Announcement Policy and IP Pool dynamically.
func ConfigureCiliumIPPool(ctx *pulumi.Context, k8sProvider *kubernetes.Provider, nodeIP pulumi.StringInput, interfaceName string) error {

	// Calculate the Blocks (CIDRs) as an Output
	blocks := nodeIP.ToStringOutput().ApplyT(func(ip string) ([]any, error) {
		parts := strings.Split(ip, ".")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid IPv4 address: %s", ip)
		}
		subnet := strings.Join(parts[:3], ".") // "192.168.1"

		var b []any
		for i := 220; i <= 230; i++ {
			b = append(b, map[string]string{
				"cidr": fmt.Sprintf("%s.%d/32", subnet, i),
			})
		}
		return b, nil
	})

	// Create IP Pool Resource
	_, err := apiextensions.NewCustomResource(ctx, "cilium-ip-pool", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cilium.io/v2alpha1"),
		Kind:       pulumi.String("CiliumLoadBalancerIPPool"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("address-pool"),
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"blocks": blocks, // Output inserted here
			},
		},
	}, pulumi.Provider(k8sProvider))
	if err != nil {
		return err
	}

	// Create L2 Policy Resource
	_, err = apiextensions.NewCustomResource(ctx, "cilium-l2-policy", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cilium.io/v2alpha1"),
		Kind:       pulumi.String("CiliumL2AnnouncementPolicy"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("l2-policy"),
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"nodeSelector": map[string]any{
					"matchExpressions": []map[string]any{
						{
							"key":      "node-role.kubernetes.io/control-plane",
							"operator": "DoesNotExist",
						},
					},
				},
				"interfaces": []string{
					fmt.Sprintf("^%s", interfaceName),
				},
				"externalIPs":     true,
				"loadBalancerIPs": true,
			},
		},
	}, pulumi.Provider(k8sProvider))

	return err
}

// PatchCiliumRBAC creates a ClusterRole and Binding to allow Cilium to manage Leases for L2 announcements.
// This is necessary if the Helm chart installation didn't include these permissions.
func PatchCiliumRBAC(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {

	// Create ClusterRole
	_, err := rbacv1.NewClusterRole(ctx, "cilium-leases-access", &rbacv1.ClusterRoleArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cilium-leases-access"),
		},
		Rules: rbacv1.PolicyRuleArray{
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("coordination.k8s.io")},
				Resources: pulumi.StringArray{pulumi.String("leases")},
				Verbs:     pulumi.StringArray{pulumi.String("get"), pulumi.String("list"), pulumi.String("watch"), pulumi.String("create"), pulumi.String("update"), pulumi.String("patch"), pulumi.String("delete")},
			},
		},
	}, pulumi.Provider(k8sProvider))
	if err != nil {
		return err
	}

	// Create ClusterRoleBinding
	_, err = rbacv1.NewClusterRoleBinding(ctx, "cilium-leases-access-binding", &rbacv1.ClusterRoleBindingArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cilium-leases-access-binding"),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      pulumi.String("cilium"),
				Namespace: pulumi.String("kube-system"),
			},
		},
		RoleRef: &rbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("ClusterRole"),
			Name:     pulumi.String("cilium-leases-access"),
		},
	}, pulumi.Provider(k8sProvider))

	return err
}
