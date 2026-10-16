package platform

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/madhank93/homelab/config"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	talos_client "github.com/pulumiverse/pulumi-talos/sdk/go/talos/client"
	talos_cluster "github.com/pulumiverse/pulumi-talos/sdk/go/talos/cluster"
	talos_machine "github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
	"go.yaml.in/yaml/v3"
)

// DeployTalosCluster provisions Talos Linux VMs on Proxmox and bootstraps the
// Kubernetes cluster. It downloads the base and GPU Talos images, generates
// machine secrets, creates the VMs listed in config/config.yml, applies per-node
// machine configs (hostname, static IP, routes), bootstraps etcd on the first
// controller, and writes talosconfig + kubeconfig to disk for the platform stack.
//
// Run `just core talos up` to apply.
//
// Changing the Talos image affects only newly created VMs; upgrade existing
// nodes with `just talos-upgrade <ip>`.
func DeployTalosCluster(ctx *pulumi.Context) error {
	// Initialize Provider & Config
	provider, err := NewProxmoxProvider(ctx)
	if err != nil {
		return err
	}
	cl := config.C.Cluster
	talos := config.V.Talos
	pveNode := config.C.Proxmox.NodeName
	clusterEndpoint := "https://" + cl.VIP + ":6443"
	factoryURL := func(schematic string) string {
		return "https://factory.talos.dev/image/" + schematic + "/" + talos.Version + "/nocloud-amd64.raw.gz"
	}

	// Base image extensions: iscsi-tools, util-linux-tools, qemu-guest-agent
	baseTalosImage, err := DownloadImage(ctx, provider, "talos-base-image", pveNode,
		factoryURL(talos.BaseSchematic),
		"talos-nocloud-amd64-base.img",
		"gz",
	)
	if err != nil {
		return err
	}

	// GPU image adds: iscsi-tools, util-linux-tools, qemu-guest-agent,
	//             nvidia-container-toolkit-production, nvidia-open-gpu-kernel-modules-production
	gpuTalosImage, err := DownloadImage(ctx, provider, "talos-gpu-image", pveNode,
		factoryURL(talos.GPUSchematic),
		"talos-nocloud-amd64-gpu.img",
		"gz",
	)
	if err != nil {
		return err
	}

	// Generate Talos Secrets
	secrets, err := talos_machine.NewSecrets(ctx, "talos-secrets", &talos_machine.SecretsArgs{
		TalosVersion: pulumi.String(talos.Version),
	})
	if err != nil {
		return err
	}

	// Transform secrets for Client package
	clientConfigInput := talos_client.GetConfigurationClientConfigurationArgs{
		CaCertificate:     secrets.ClientConfiguration.CaCertificate(),
		ClientCertificate: secrets.ClientConfiguration.ClientCertificate(),
		ClientKey:         secrets.ClientConfiguration.ClientKey(),
	}

	// Generate Client Config (Talos Config)
	clientConfig := talos_client.GetConfigurationOutput(ctx, talos_client.GetConfigurationOutputArgs{
		ClusterName:         pulumi.String(cl.Name),
		ClientConfiguration: clientConfigInput,
		Endpoints:           pulumi.StringArray{pulumi.String(cl.VIP)},
	})

	ctx.Export("talosconfig", clientConfig.TalosConfig())

	// Write Talosconfig to local file
	clientConfig.TalosConfig().ApplyT(func(tc string) (any, error) {
		err := os.WriteFile("talosconfig", []byte(tc), 0o600)
		return nil, err
	})

	// Patches for DHCP and VIP
	cpPatch := fmt.Sprintf(`machine:
  network:
    interfaces:
      - deviceSelector:
          physical: true
        dhcp: true
        vip:
          ip: %s
`, cl.VIP)

	workerPatch := fmt.Sprintf(`machine:
  nodeLabels:
    "node.longhorn.io/create-default-disk": "config"
  network:
    interfaces:
      - deviceSelector:
          physical: true
        dhcp: true
  kubelet:
    nodeIP:
      validSubnets:
        - %s
  kernel:
    modules:
      - name: nbd
      - name: iscsi_tcp
      - name: iscsi_generic
      - name: configfs
`, cl.Subnet)

	// GPU worker = worker patch + this; Talos appends the module list.
	// Note: The nvidia-container-toolkit-production Talos extension configures
	// containerd automatically — no machine.files drop-in needed.
	// Talos v1.10+ restricts machine.files writes to /var; /etc/cri/conf.d/ is not allowed.
	gpuWorkerPatch := `machine:
  nodeTaints:
    dedicated: "ai:NoSchedule"
  kernel:
    modules:
      - name: nvidia
      - name: nvidia_uvm
      - name: nvidia_drm
      - name: nvidia_modeset
`

	// install.disk is required: these nodes boot with talos.platform=metal, and the
	// metal installer refuses to run an upgrade without it. sda is the 215 GB virtio
	// system disk on every node; sdb on the workers is the iSCSI volume Longhorn uses.
	// install.image is deliberately unset — deprecated in favour of Image Factory,
	// and `just talos-upgrade` passes the factory installer explicitly.
	basePatch := `cluster:
  network:
    cni:
      name: none
  proxy:
    disabled: true
machine:
  install:
    disk: /dev/sda
`

	// Generate Configs
	cpConfig := talos_machine.GetConfigurationOutput(ctx, talos_machine.GetConfigurationOutputArgs{
		ClusterName:     pulumi.String(cl.Name),
		MachineType:     pulumi.String("controlplane"),
		ClusterEndpoint: pulumi.String(clusterEndpoint),
		MachineSecrets:  secrets.MachineSecrets,
		TalosVersion:    pulumi.String(talos.Version),
		ConfigPatches:   pulumi.StringArray{pulumi.String(basePatch), pulumi.String(cpPatch)},
	})

	workerConfig := talos_machine.GetConfigurationOutput(ctx, talos_machine.GetConfigurationOutputArgs{
		ClusterName:     pulumi.String(cl.Name),
		MachineType:     pulumi.String("worker"),
		ClusterEndpoint: pulumi.String(clusterEndpoint),
		MachineSecrets:  secrets.MachineSecrets,
		TalosVersion:    pulumi.String(talos.Version),
		ConfigPatches:   pulumi.StringArray{pulumi.String(basePatch), pulumi.String(workerPatch)},
	})

	gpuConfig := talos_machine.GetConfigurationOutput(ctx, talos_machine.GetConfigurationOutputArgs{
		ClusterName:     pulumi.String(cl.Name),
		MachineType:     pulumi.String("worker"),
		ClusterEndpoint: pulumi.String(clusterEndpoint),
		MachineSecrets:  secrets.MachineSecrets,
		TalosVersion:    pulumi.String(talos.Version),
		ConfigPatches:   pulumi.StringArray{pulumi.String(basePatch), pulumi.String(workerPatch), pulumi.String(gpuWorkerPatch)},
	})

	// Export sanitized configs
	ctx.Export("cp-config", cpConfig.MachineConfiguration())
	ctx.Export("worker-config", workerConfig.MachineConfiguration())

	var cpConfigApplies []pulumi.Resource
	var controllerIP pulumi.String

	for _, node := range cl.Nodes {
		// Select Image Based on GPU
		var imageID pulumi.IDOutput
		if node.HasGPU() {
			imageID = gpuTalosImage.ID()
		} else {
			imageID = baseTalosImage.ID()
		}

		// Create VM
		vmRes, err := NewProxmoxVM(ctx, provider, pveNode, node, imageID)
		if err != nil {
			return err
		}

		// Use Static IP
		nodeIP := pulumi.String(node.IP)

		ctx.Export(node.Name+"-ip", nodeIP)

		if controllerIP == "" && node.Role == "control" {
			controllerIP = nodeIP
		}

		// Select Config Base
		var baseConfig pulumi.StringOutput
		if node.Role == "control" {
			baseConfig = cpConfig.MachineConfiguration()
		} else if node.HasGPU() {
			baseConfig = gpuConfig.MachineConfiguration()
		} else {
			baseConfig = workerConfig.MachineConfiguration()
		}

		// Patch Hostname and Sanitize
		finalConfig := baseConfig.ApplyT(func(c string) (string, error) {
			return patchTalosConfig(c, node.Name, node.IP, cl)
		}).(pulumi.StringOutput)

		// Dynamically discovered IP from QEMU agent for bootstrapping.
		// Falls back to static node IP if the VM is stopped (no QEMU agent response)
		// so that a stopped VM doesn't block the Pulumi preview for other nodes.
		dynamicEndpoint := vmRes.Ipv4Addresses.ApplyT(func(ips [][]string) (string, error) {
			for _, net := range ips {
				for _, ip := range net {
					// Filter out loopback, link-local APIPA, and IPv6 addresses
					if ip != "127.0.0.1" && ip != "" && !strings.HasPrefix(ip, "169.254.") && !strings.Contains(ip, ":") {
						return ip, nil
					}
				}
			}
			// VM is stopped or QEMU agent not yet ready — fall back to static IP.
			// ConfigurationApply will fail gracefully if the node is truly unreachable.
			return node.IP, nil
		}).(pulumi.StringOutput)

		// Client Config construct
		clientConfigArgs := talos_machine.ClientConfigurationArgs{
			CaCertificate:     secrets.ClientConfiguration.CaCertificate(),
			ClientCertificate: secrets.ClientConfiguration.ClientCertificate(),
			ClientKey:         secrets.ClientConfiguration.ClientKey(),
		}

		// Apply Config In-Band
		cfgApply, err := talos_machine.NewConfigurationApply(ctx, node.Name+"-config-apply", &talos_machine.ConfigurationApplyArgs{
			ClientConfiguration:       clientConfigArgs,
			MachineConfigurationInput: finalConfig,
			Node:                      nodeIP,
			Endpoint:                  dynamicEndpoint,
		}, pulumi.DependsOn([]pulumi.Resource{vmRes}))
		if err != nil {
			return err
		}

		if node.Role == "control" {
			cpConfigApplies = append(cpConfigApplies, cfgApply)
		}
	}

	// Bootstrap
	bootstrapClientConfig := talos_machine.ClientConfigurationArgs{
		CaCertificate:     secrets.ClientConfiguration.CaCertificate(),
		ClientCertificate: secrets.ClientConfiguration.ClientCertificate(),
		ClientKey:         secrets.ClientConfiguration.ClientKey(),
	}

	// Bootstrap against the FIRST Controller Node IP (k8s-controller1) to ensure safe leader election
	bootstrap, err := talos_machine.NewBootstrap(ctx, "bootstrap", &talos_machine.BootstrapArgs{
		ClientConfiguration: bootstrapClientConfig,
		Node:                controllerIP,
	}, pulumi.DependsOn(cpConfigApplies))
	if err != nil {
		return err
	}

	// Kubeconfig
	clusterClientConfigInput := talos_cluster.KubeconfigClientConfigurationArgs{
		CaCertificate:     secrets.ClientConfiguration.CaCertificate(),
		ClientCertificate: secrets.ClientConfiguration.ClientCertificate(),
		ClientKey:         secrets.ClientConfiguration.ClientKey(),
	}

	kubeconfigRes, err := talos_cluster.NewKubeconfig(ctx, "kubeconfig", &talos_cluster.KubeconfigArgs{
		ClientConfiguration: clusterClientConfigInput,
		Node:                pulumi.String(cl.VIP),
	}, pulumi.DependsOn([]pulumi.Resource{bootstrap}))
	if err != nil {
		return err
	}

	// Write Kubeconfig to local file — read by the platform stack via DeployPlatform
	kubeconfigRes.KubeconfigRaw.ApplyT(func(kc string) (any, error) {
		err := os.WriteFile("kubeconfig", []byte(kc), 0o600)
		return nil, err
	})

	ctx.Export("kubeconfig", kubeconfigRes.KubeconfigRaw)

	return nil
}

// patchTalosConfig applies per-node patches to a raw Talos machine config YAML.
//
// It performs three transformations:
//  1. Drops the HostnameConfig document (auto-generated; we set hostname explicitly).
//  2. Removes machine.install (handled by cloud-init; not needed after first boot).
//  3. Sets machine.network.hostname, converts DHCP to a static address (hostname/24),
//     adds a default gateway route, and adds a static route for the NetBird subnet
//     (100.109.0.0/16 via worker1) so return traffic can reach the WireGuard tunnel.
func patchTalosConfig(rawConfig, hostname, ip string, cl config.ClusterConfig) (string, error) {
	prefix := cl.Subnet[strings.Index(cl.Subnet, "/"):]
	reader := strings.NewReader(rawConfig)
	decoder := yaml.NewDecoder(reader)
	var documents []map[string]any

	for {
		var doc map[string]any
		err := decoder.Decode(&doc)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		if kind, ok := doc["kind"].(string); ok && kind == "HostnameConfig" {
			continue
		}

		if _, hasMachine := doc["machine"]; hasMachine {
			if machineMap, ok := doc["machine"].(map[string]any); ok {
				if _, hasNetwork := machineMap["network"]; !hasNetwork {
					machineMap["network"] = make(map[string]any)
				}
				if networkMap, ok := machineMap["network"].(map[string]any); ok {
					networkMap["hostname"] = hostname
					if interfaces, ok := networkMap["interfaces"].([]any); ok && len(interfaces) > 0 {
						if iface, ok := interfaces[0].(map[string]any); ok {
							iface["dhcp"] = false
							iface["addresses"] = []string{ip + prefix}
							iface["routes"] = []any{
								map[string]any{
									"gateway": cl.Router,
								},
								// Static route: return traffic to Bifrost's NetBird subnet
								// must go via k8s-routing-peer (worker1). Cilium's eBPF
								// bypasses iptables MASQUERADE, so a kernel route is required
								// on every node so SYN-ACKs reach worker1's wt0 tunnel.
								map[string]any{
									"network": cl.NetbirdCIDR,
									"gateway": cl.NetbirdRouter,
									"metric":  1024,
								},
							}
						}
					}
					networkMap["nameservers"] = cl.Nameservers
				}
			}
		}
		documents = append(documents, doc)
	}

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	for _, doc := range documents {
		if err := encoder.Encode(doc); err != nil {
			return "", err
		}
	}
	encoder.Close()
	return buf.String(), nil
}
