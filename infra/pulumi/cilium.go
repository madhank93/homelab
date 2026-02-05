package main

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

func InstallCilium(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {

	// Define Cilium Helm Chart
	_, err := helm.NewRelease(ctx, "cilium", &helm.ReleaseArgs{
		Chart:   pulumi.String("cilium"),
		Version: pulumi.String("1.16.6"), // Latest Stable
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
			"k8sServiceHost": pulumi.String("192.168.1.100"), // VIP
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
		},
	}, pulumi.Provider(k8sProvider))

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

	// Create GatewayClass Resource
	_, err = apiextensions.NewCustomResource(ctx, "cilium-gateway-class", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("gateway.networking.k8s.io/v1"),
		Kind:       pulumi.String("GatewayClass"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cilium"),
		},
		OtherFields: map[string]interface{}{
			"spec": map[string]interface{}{
				"controllerName": "io.cilium/gateway-controller",
			},
		},
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{crds}))
	if err != nil {
		return err
	}

	// Create Gateway Resource (Depends on GatewayClass)
	_, err = apiextensions.NewCustomResource(ctx, "cilium-gateway", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("gateway.networking.k8s.io/v1"),
		Kind:       pulumi.String("Gateway"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("homelab-gateway"),
			Namespace: pulumi.String("kube-system"),
		},
		OtherFields: map[string]interface{}{
			"spec": map[string]interface{}{
				"gatewayClassName": "cilium",
				"listeners": []map[string]interface{}{
					{
						"name":     "http",
						"protocol": "HTTP",
						"port":     80,
						"allowedRoutes": map[string]interface{}{
							"namespaces": map[string]interface{}{
								"from": "All",
							},
						},
					},
					{
						"name":     "https",
						"protocol": "TLS", // Must be TLS for Passthrough mode
						"port":     443,
						"hostname": "*.madhan.app", // Match any subddomain
						"tls": map[string]interface{}{
							"mode": "Passthrough", // Let ArgoCD handle TLS, or "Terminate" if we had certs
							// using Passthrough for now to keep it simple with self-signed upstream
						},
						"allowedRoutes": map[string]any{
							"namespaces": map[string]interface{}{
								"from": "All",
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
	blocks := nodeIP.ToStringOutput().ApplyT(func(ip string) ([]interface{}, error) {
		parts := strings.Split(ip, ".")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid IPv4 address: %s", ip)
		}
		subnet := strings.Join(parts[:3], ".") // "192.168.1"

		var b []interface{}
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
