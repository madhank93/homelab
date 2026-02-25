package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	talos_client "github.com/pulumiverse/pulumi-talos/sdk/go/talos/client"
	talos_cluster "github.com/pulumiverse/pulumi-talos/sdk/go/talos/cluster"
	talos_machine "github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
	"go.yaml.in/yaml/v3"
)

const (
	talosVersion    = "v1.12.4"
	clusterEndpoint = "https://192.168.1.210:6443" // VIP
	vipIP           = "192.168.1.210"
)

func DeployTalosCluster(ctx *pulumi.Context) error {
	// Initialize Provider & Config
	provider, cfg, err := NewProxmoxProvider(ctx)
	if err != nil {
		return err
	}

	// Download Base Talos Image (without Nvidia extensions)
	// Schematic ID: 88d1f7a5c4f1d3aba7df787c448c1d3d008ed29cfb34af53fa0df4336a56040b
	// Extensions: iscsi-tools, util-linux-tools, qemu-guest-agent
	baseTalosImage, err := DownloadImage(ctx, provider, "talos-base-image", cfg.NodeName,
		"https://factory.talos.dev/image/88d1f7a5c4f1d3aba7df787c448c1d3d008ed29cfb34af53fa0df4336a56040b/v1.12.4/nocloud-amd64.raw.gz",
		"talos-nocloud-amd64-base.img",
		"gz",
	)
	if err != nil {
		return err
	}

	// Download GPU Talos Image (with Nvidia extensions)
	// Schematic ID: 901b9afcf2f7eda57991690fc5ca00414740cc4ee4ad516109bcc58beff1b829
	// Extensions: iscsi-tools, util-linux-tools, qemu-guest-agent, nvidia-container-toolkit, nvidia-open-gpu-kernel-modules
	gpuTalosImage, err := DownloadImage(ctx, provider, "talos-gpu-image", cfg.NodeName,
		"https://factory.talos.dev/image/901b9afcf2f7eda57991690fc5ca00414740cc4ee4ad516109bcc58beff1b829/v1.12.4/nocloud-amd64.raw.gz",
		"talos-nocloud-amd64-gpu.img",
		"gz",
	)
	if err != nil {
		return err
	}

	// Generate Talos Secrets
	secrets, err := talos_machine.NewSecrets(ctx, "talos-secrets", &talos_machine.SecretsArgs{
		TalosVersion: pulumi.String(talosVersion),
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
		ClusterName:         pulumi.String("talos-cluster"),
		ClientConfiguration: clientConfigInput,
		Endpoints:           pulumi.StringArray{pulumi.String(vipIP)},
	})

	ctx.Export("talosconfig", clientConfig.TalosConfig())

	// Write Talosconfig to local file
	clientConfig.TalosConfig().ApplyT(func(tc string) (any, error) {
		err := os.WriteFile("talosconfig", []byte(tc), 0o600)
		return nil, err
	})

	// Define nodes
	nodes := []NodeConfig{
		// Control Plane (3 Nodes) - High Priority
		{Name: "k8s-controller1", IP: "192.168.1.211", Role: "control", Cores: 4, Memory: 6144, DiskSize: 30, CpuUnits: 1024, Balloon: 4096},
		{Name: "k8s-controller2", IP: "192.168.1.212", Role: "control", Cores: 4, Memory: 6144, DiskSize: 30, CpuUnits: 1024, Balloon: 4096},
		{Name: "k8s-controller3", IP: "192.168.1.213", Role: "control", Cores: 4, Memory: 6144, DiskSize: 30, CpuUnits: 1024, Balloon: 4096},
		// Workers (4 Nodes) - Standard Priority
		{Name: "k8s-worker1", IP: "192.168.1.221", Role: "worker", Cores: 4, Memory: 6144, DiskSize: 125, CpuUnits: 100, Balloon: 2048},
		{Name: "k8s-worker2", IP: "192.168.1.222", Role: "worker", Cores: 4, Memory: 6144, DiskSize: 125, CpuUnits: 100, Balloon: 2048},
		{Name: "k8s-worker3", IP: "192.168.1.223", Role: "worker", Cores: 4, Memory: 6144, DiskSize: 125, CpuUnits: 100, Balloon: 2048},
		// Worker 4 with GPU
		{
			Name: "k8s-worker4", IP: "192.168.1.224", Role: "worker", Cores: 4, Memory: 6144, DiskSize: 125,
			HasGPU: true, PcieIDs: []string{"0000:28:00.0"},
			CpuUnits: 100, Balloon: 2048,
		},
	}

	// Patches for DHCP and VIP
	cpPatch := fmt.Sprintf(`machine:
  network:
    interfaces:
      - deviceSelector:
          physical: true
        dhcp: true
        vip:
          ip: %s
`, vipIP)

	workerPatch := `machine:
  nodeLabels:
    "node.longhorn.io/create-default-disk": "config"
  network:
    interfaces:
      - deviceSelector:
          physical: true
        dhcp: true
  kernel:
    modules:
      - name: nbd
      - name: iscsi_tcp
      - name: iscsi_generic
      - name: configfs
`

	// GPU Worker Patch - Includes worker networking/modules + Nvidia modules
	gpuWorkerPatch := `machine:
  nodeLabels:
    "node.longhorn.io/create-default-disk": "config"
  network:
    interfaces:
      - deviceSelector:
          physical: true
        dhcp: true
  kernel:
    modules:
      - name: nbd
      - name: iscsi_tcp
      - name: iscsi_generic
      - name: configfs
      - name: nvidia
      - name: nvidia_uvm
      - name: nvidia_drm
      - name: nvidia_modeset
`

	basePatch := `cluster:
  network:
    cni:
      name: none
  proxy:
    disabled: true
`

	// Generate Configs
	cpConfig := talos_machine.GetConfigurationOutput(ctx, talos_machine.GetConfigurationOutputArgs{
		ClusterName:     pulumi.String("talos-cluster"),
		MachineType:     pulumi.String("controlplane"),
		ClusterEndpoint: pulumi.String(clusterEndpoint),
		MachineSecrets:  secrets.MachineSecrets,
		TalosVersion:    pulumi.String(talosVersion),
		ConfigPatches:   pulumi.StringArray{pulumi.String(basePatch), pulumi.String(cpPatch)},
	})

	workerConfig := talos_machine.GetConfigurationOutput(ctx, talos_machine.GetConfigurationOutputArgs{
		ClusterName:     pulumi.String("talos-cluster"),
		MachineType:     pulumi.String("worker"),
		ClusterEndpoint: pulumi.String(clusterEndpoint),
		MachineSecrets:  secrets.MachineSecrets,
		TalosVersion:    pulumi.String(talosVersion),
		ConfigPatches:   pulumi.StringArray{pulumi.String(basePatch), pulumi.String(workerPatch)},
	})

	gpuConfig := talos_machine.GetConfigurationOutput(ctx, talos_machine.GetConfigurationOutputArgs{
		ClusterName:     pulumi.String("talos-cluster"),
		MachineType:     pulumi.String("worker"),
		ClusterEndpoint: pulumi.String(clusterEndpoint),
		MachineSecrets:  secrets.MachineSecrets,
		TalosVersion:    pulumi.String(talosVersion),
		ConfigPatches:   pulumi.StringArray{pulumi.String(basePatch), pulumi.String(gpuWorkerPatch)},
	})

	// Export sanitized configs
	ctx.Export("cp-config", cpConfig.MachineConfiguration())
	ctx.Export("worker-config", workerConfig.MachineConfiguration())

	var cpConfigApplies []pulumi.Resource
	var controllerIP pulumi.String

	for _, node := range nodes {
		// Select Image Based on GPU
		var imageID pulumi.IDOutput
		if node.HasGPU {
			imageID = gpuTalosImage.ID()
		} else {
			imageID = baseTalosImage.ID()
		}

		// Create VM
		vmRes, err := NewProxmoxVM(ctx, provider, cfg.NodeName, node, imageID)
		if err != nil {
			return err
		}

		// Use Static IP
		nodeIP := pulumi.String(node.IP)

		ctx.Export(node.Name+"-ip", nodeIP)

		if node.Name == "k8s-controller1" {
			controllerIP = nodeIP
		}

		// Select Config Base
		var baseConfig pulumi.StringOutput
		if node.Role == "control" {
			baseConfig = cpConfig.MachineConfiguration()
		} else if node.HasGPU {
			baseConfig = gpuConfig.MachineConfiguration()
		} else {
			baseConfig = workerConfig.MachineConfiguration()
		}

		// Patch Hostname and Sanitize
		finalConfig := baseConfig.ApplyT(func(c string) (string, error) {
			return patchTalosConfig(c, node.Name, node.IP)
		}).(pulumi.StringOutput)

		// Dynamically discovered IP from QEMU agent for bootstrapping
		dynamicEndpoint := vmRes.Ipv4Addresses.ApplyT(func(ips [][]string) (string, error) {
			for _, net := range ips {
				for _, ip := range net {
					// Filter out loopback, link-local APIPA, and IPv6 addresses
					if ip != "127.0.0.1" && ip != "" && !strings.HasPrefix(ip, "169.254.") && !strings.Contains(ip, ":") {
						return ip, nil
					}
				}
			}
			return "", fmt.Errorf("no valid IPv4 address reported by QEMU agent yet")
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
		Node:                pulumi.String(vipIP),
	}, pulumi.DependsOn([]pulumi.Resource{bootstrap}))
	if err != nil {
		return err
	}

	// Write Kubeconfig to local file
	kubeconfigRes.KubeconfigRaw.ApplyT(func(kc string) (any, error) {
		err := os.WriteFile("kubeconfig", []byte(kc), 0o600)
		return nil, err
	})

	ctx.Export("kubeconfig", kubeconfigRes.KubeconfigRaw)

	// Initialize Kubernetes Provider (for CNI & GitOps)
	// We depend on the Bootstrap having finished and Kubeconfig being generated.
	k8sProvider, err := kubernetes.NewProvider(ctx, "k8s-provider", &kubernetes.ProviderArgs{
		Kubeconfig: kubeconfigRes.KubeconfigRaw,
	}, pulumi.DependsOn([]pulumi.Resource{kubeconfigRes, bootstrap}))
	if err != nil {
		return err
	}

	// Install Cilium (CNI) - Required for Nodes to become Ready
	if err := InstallCilium(ctx, k8sProvider); err != nil {
		return err
	}

	// Install Gateway (Gateway API)
	if err := InstallGateway(ctx, k8sProvider); err != nil {
		return err
	}

	// Install cert-manager (TLS certificate management)
	// Requires cert-manager/cloudflare-api-token Secret — run `just create-secrets` first.
	if err := InstallCertManager(ctx, k8sProvider); err != nil {
		return err
	}

	// Install ArgoCD (GitOps)
	if err := InstallArgoCD(ctx, k8sProvider); err != nil {
		return err
	}

	// Configure Dynamic Cilium IP Pool
	// Uses the first controller's IP to determine subnet and "eth0" as interface.
	if err := ConfigureCiliumIPPool(ctx, k8sProvider, controllerIP, "(eth0|ens.*|enp.*)"); err != nil {
		return err
	}

	// Patch Cilium RBAC for Leases (L2 Announcement Fix)
	if err := PatchCiliumRBAC(ctx, k8sProvider); err != nil {
		return err
	}

	return nil
}

// patchTalosConfig removes HostnameConfig, machine.install, and sets the hostname.
func patchTalosConfig(rawConfig, hostname, ip string) (string, error) {
	reader := strings.NewReader(rawConfig)
	decoder := yaml.NewDecoder(reader)
	var documents []map[string]any

	for {
		var doc map[string]any
		err := decoder.Decode(&doc)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}

		if kind, ok := doc["kind"].(string); ok && kind == "HostnameConfig" {
			continue
		}

		if _, hasMachine := doc["machine"]; hasMachine {
			if machineMap, ok := doc["machine"].(map[string]any); ok {
				delete(machineMap, "install")
				if _, hasNetwork := machineMap["network"]; !hasNetwork {
					machineMap["network"] = make(map[string]any)
				}
				if networkMap, ok := machineMap["network"].(map[string]any); ok {
					networkMap["hostname"] = hostname
					if interfaces, ok := networkMap["interfaces"].([]any); ok && len(interfaces) > 0 {
						if iface, ok := interfaces[0].(map[string]any); ok {
							iface["dhcp"] = false
							iface["addresses"] = []string{fmt.Sprintf("%s/24", ip)}
							iface["routes"] = []any{
								map[string]any{
									"gateway": "192.168.1.254",
								},
							}
						}
					}
					networkMap["nameservers"] = []string{"1.1.1.1", "192.168.1.254"}
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
