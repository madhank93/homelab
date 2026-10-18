package platform

import (
	"fmt"
	"net/netip"

	"github.com/madhank93/homelab/config"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// InstallCilium installs Cilium as the cluster CNI and kube-proxy replacement,
// plus an HTTPRoute for hubble.madhan.app.
//
// devices must never include wt0: it is NOARP/POINTOPOINT, and Cilium's TC BPF
// silently drops non-Ethernet frames. NetBird traffic reaches Cilium by kernel
// IP forwarding to the LAN NIC of another node.
func InstallCilium(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {
	// Held on 1.18.x: 1.19.x breaks host TCP on eth0 nodes (cilium/cilium#44430,
	// #46010). Fixed in 1.20.0, which needs its own change window — a bad
	// upgrade takes out the Talos API too.
	cilium := config.HelmChart("cilium")
	ciliumChart, err := helm.NewRelease(ctx, "cilium", &helm.ReleaseArgs{
		Chart:   pulumi.String("cilium"),
		Version: pulumi.String(cilium.Version),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String(cilium.Repo),
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
			"k8sServiceHost": pulumi.String(config.C.Cluster.VIP),
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
			"l2announcements": pulumi.Map{
				"enabled": pulumi.Bool(true),
			},
			"gatewayAPI": pulumi.Map{
				"enabled": pulumi.Bool(true),
				"secretsNamespace": pulumi.Map{
					// Cilium copies the Gateway's TLS secret into cilium-secrets for Envoy
					// SDS. It must own that copy: cilium-secrets is Cilium's own namespace,
					// so a foreign secret parked there gets removed, and if cert-manager is
					// what recreates it, every removal costs a fresh ACME order. That loop
					// exhausted Let's Encrypt's 5-per-168h duplicate-certificate limit and
					// took down HTTPS on every route. One Certificate in kube-system, one
					// owner for the copy.
					"sync": pulumi.Bool(true),
				},
			},
			// ens18 is the VirtIO NIC under predictable naming, eth0 under classic
			// naming; nodes have carried both across Talos upgrades. Never add wt0
			// (see InstallCilium doc).
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
					{"name": config.C.Gateway.Name, "namespace": config.C.Gateway.Namespace},
				},
				"hostnames": []string{config.C.Host("hubble")},
				"rules": []map[string]any{
					{
						"matches":     []map[string]any{{"path": map[string]any{"type": "PathPrefix", "value": "/"}}},
						"backendRefs": []map[string]any{{"name": "hubble-ui", "port": 80}},
					},
				},
			},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{ciliumChart}))
	if err != nil {
		return err
	}

	return denyServiceExternalIPs(ctx, k8sProvider)
}

// denyServiceExternalIPs rejects Services that set spec.externalIPs, which would
// let any Service author hijack traffic to an arbitrary IP (CVE-2020-8554).
// Cilium cannot disable this itself: kubeProxyReplacement forces it on.
func denyServiceExternalIPs(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {
	policy, err := apiextensions.NewCustomResource(ctx, "deny-service-external-ips", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("admissionregistration.k8s.io/v1"),
		Kind:       pulumi.String("ValidatingAdmissionPolicy"),
		Metadata:   &metav1.ObjectMetaArgs{Name: pulumi.String("deny-service-external-ips")},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"failurePolicy": "Fail",
				"matchConstraints": map[string]any{
					"resourceRules": []map[string]any{{
						"apiGroups":   []string{""},
						"apiVersions": []string{"v1"},
						"operations":  []string{"CREATE", "UPDATE"},
						"resources":   []string{"services"},
					}},
				},
				"validations": []map[string]any{{
					"expression": "!has(object.spec.externalIPs) || size(object.spec.externalIPs) == 0",
					"message":    "spec.externalIPs is not allowed; use a LoadBalancer Service",
				}},
			},
		},
	}, pulumi.Provider(k8sProvider))
	if err != nil {
		return err
	}

	_, err = apiextensions.NewCustomResource(ctx, "deny-service-external-ips-binding", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("admissionregistration.k8s.io/v1"),
		Kind:       pulumi.String("ValidatingAdmissionPolicyBinding"),
		Metadata:   &metav1.ObjectMetaArgs{Name: pulumi.String("deny-service-external-ips")},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"policyName":        "deny-service-external-ips",
				"validationActions": []string{"Deny"},
			},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{policy}))

	return err
}

// InstallGateway creates the shared Gateway resource for the cluster
func InstallGateway(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {

	// Gateway API CRDs (experimental channel v1.2.1) — required by Cilium 1.16+.
	//
	// Vendored rather than fetched from the GitHub release URL: that fetch ran on
	// every apply, so a rate limit or a blip failed the whole stack with an opaque
	// "cannot unmarshal string into Go value of type map[string]interface{}" —
	// GitHub's error page parsed as a bare string. The version is now pinned in
	// git like every other dependency. Re-vendor to upgrade; do not point this
	// back at a URL.
	crds, err := yaml.NewConfigFile(ctx, "gateway-api-crds", &yaml.ConfigFileArgs{
		File: "platform/manifests/gateway-api-v1.2.1-experimental-install.yaml",
	}, pulumi.Provider(k8sProvider))
	if err != nil {
		return err
	}

	// Create Gateway Resource (Depends on GatewayClass from Cilium Helm Chart + CRDs)
	_, err = apiextensions.NewCustomResource(ctx, "cilium-gateway", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("gateway.networking.k8s.io/v1"),
		Kind:       pulumi.String("Gateway"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(config.C.Gateway.Name),
			Namespace: pulumi.String(config.C.Gateway.Namespace),
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

// ConfigureCiliumIPPool creates the L2 announcement policy and the LB IP pool
// (one /32 block per address in config cluster.lb_pool).
func ConfigureCiliumIPPool(ctx *pulumi.Context, k8sProvider *kubernetes.Provider, interfaceName string) error {
	start, err := netip.ParseAddr(config.C.Cluster.LBPool.Start)
	if err != nil {
		return err
	}
	stop, err := netip.ParseAddr(config.C.Cluster.LBPool.Stop)
	if err != nil {
		return err
	}
	var blocks []any
	for ip := start; ip.Compare(stop) <= 0; ip = ip.Next() {
		blocks = append(blocks, map[string]string{"cidr": ip.String() + "/32"})
	}

	// Create IP Pool Resource
	_, err = apiextensions.NewCustomResource(ctx, "cilium-ip-pool", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cilium.io/v2alpha1"),
		Kind:       pulumi.String("CiliumLoadBalancerIPPool"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("address-pool"),
		},
		OtherFields: map[string]any{
			"spec": map[string]any{
				"blocks": blocks,
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
				"externalIPs":     false,
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
