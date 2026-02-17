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
	talosVersion    = "v1.9.3"
	clusterEndpoint = "https://192.168.1.100:6443" // VIP
	vipIP           = "192.168.1.100"
)

func DeployTalosCluster(ctx *pulumi.Context) error {
	// Initialize Provider & Config
	provider, cfg, err := NewProxmoxProvider(ctx)
	if err != nil {
		return err
	}

	// Download Talos Image with System Extensions
	// Schematic ID: e187c9b90f773cd8c84e5a3265c5554ee787b2fe67b508d9f955e90e7ae8c96c
	// Extensions: iscsi-tools, util-linux-tools, qemu-guest-agent
	// See: docs/talos-upgrade-guide.md for schematic configuration
	talosImage, err := DownloadImage(ctx, provider, "talos-image", cfg.NodeName,
		"https://factory.talos.dev/image/e187c9b90f773cd8c84e5a3265c5554ee787b2fe67b508d9f955e90e7ae8c96c/v1.9.3/nocloud-amd64.raw.gz",
		"talos-nocloud-amd64.img",
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
		{Name: "k8s-controller1", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30, CpuUnits: 1024, Balloon: 4096},
		{Name: "k8s-controller2", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30, CpuUnits: 1024, Balloon: 4096},
		{Name: "k8s-controller3", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30, CpuUnits: 1024, Balloon: 4096},
		// Workers (4 Nodes) - Standard Priority
		{Name: "k8s-worker1", Role: "worker", Cores: 4, Memory: 6144, DiskSize: 125, CpuUnits: 100, Balloon: 2048},
		{Name: "k8s-worker2", Role: "worker", Cores: 4, Memory: 6144, DiskSize: 125, CpuUnits: 100, Balloon: 2048},
		{Name: "k8s-worker3", Role: "worker", Cores: 4, Memory: 6144, DiskSize: 125, CpuUnits: 100, Balloon: 2048},
		// Worker 4 with GPU
		{
			Name: "k8s-worker4", Role: "worker", Cores: 4, Memory: 6144, DiskSize: 125,
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

	// Export sanitized configs
	ctx.Export("cp-config", cpConfig.MachineConfiguration())
	ctx.Export("worker-config", workerConfig.MachineConfiguration())

	var cpConfigApplies []pulumi.Resource
	var controllerIP pulumi.StringOutput

	for _, node := range nodes {
		// Create VM
		vmRes, err := NewProxmoxVM(ctx, provider, cfg.NodeName, node, talosImage.ID())
		if err != nil {
			return err
		}

		// Capture IP from Agent
		nodeIP := vmRes.Ipv4Addresses.ApplyT(func(ips [][]string) (string, error) {
			for _, netInterface := range ips {
				for _, ip := range netInterface {
					if strings.HasPrefix(ip, "192.168.") {
						return ip, nil
					}
				}
			}
			return "", fmt.Errorf("waiting for ip address (run pulumi refresh if vm is running)")
		}).(pulumi.StringOutput)

		ctx.Export(node.Name+"-ip", nodeIP)

		if node.Name == "k8s-controller1" {
			controllerIP = nodeIP
		}

		// Select Config Base
		var baseConfig pulumi.StringOutput
		if node.Role == "control" {
			baseConfig = cpConfig.MachineConfiguration()
		} else {
			baseConfig = workerConfig.MachineConfiguration()
		}

		// Patch Hostname and Sanitize
		finalConfig := baseConfig.ApplyT(func(c string) (string, error) {
			return patchTalosConfig(c, node.Name)
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
func patchTalosConfig(rawConfig, hostname string) (string, error) {
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
