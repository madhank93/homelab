package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve/download"
	"github.com/muhlba91/pulumi-proxmoxve/sdk/v7/go/proxmoxve/vm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	talos_client "github.com/pulumiverse/pulumi-talos/sdk/go/talos/client"
	talos_cluster "github.com/pulumiverse/pulumi-talos/sdk/go/talos/cluster"
	talos_machine "github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
	"gopkg.in/yaml.v3"
)

// Proxmox-specific config structs
type ProxmoxClusterConfig struct {
	Username string `koanf:"username"`
	Endpoint string `koanf:"endpoint"`
	NodeName string `koanf:"nodename"`
	ImageUrl string `koanf:"image_url"`
}

type NodeConfig struct {
	Name       string
	Role       string
	Cores      int
	Memory     int
	DiskSize   int
	HasGPU     bool
	PcieIDs    []string
	IsTemplate bool
	TemplateID pulumi.IntInput
}

const (
	talosVersion    = "v1.9.3"
	clusterEndpoint = "https://192.168.1.100:6443" // VIP
	vipIP           = "192.168.1.100"
)

func DeployProxmox(ctx *pulumi.Context) error {
	// Generic load
	var cfg ProxmoxClusterConfig
	if err := LoadConfig("proxmox", &cfg); err != nil {
		return err
	}

	// Access password from global k
	password := k.String("PROXMOX_PASSWORD")
	if password == "" {
		return fmt.Errorf("PROXMOX_PASSWORD not found in .env")
	}

	// Initialize provider
	provider, err := proxmoxve.NewProvider(ctx, "proxmoxve", &proxmoxve.ProviderArgs{
		Endpoint: pulumi.String(cfg.Endpoint),
		Username: pulumi.String(cfg.Username),
		Password: pulumi.String(password),
		Insecure: pulumi.Bool(true),
	})
	if err != nil {
		return err
	}

	// Download Talos Image (Disk Image)
	// Using factory Raw image for v1.9.3 with qemu-guest-agent
	talosImage, err := download.NewFile(ctx, "talos-image", &download.FileArgs{
		ContentType:            pulumi.String("iso"),
		DatastoreId:            pulumi.String("local"),
		NodeName:               pulumi.String(cfg.NodeName),
		Url:                    pulumi.String("https://factory.talos.dev/image/ce4c980550dd2ab1b17bbf2b08801c7eb59418eafe8f279833297925d67c7515/v1.9.3/nocloud-amd64.raw.gz"),
		FileName:               pulumi.String("talos-nocloud-amd64.img"),
		DecompressionAlgorithm: pulumi.String("gz"),
		Overwrite:              pulumi.Bool(false),
	}, pulumi.Provider(provider))
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

	// Generate Client Config (Talos Config) -> Talking to VIP
	clientConfig := talos_client.GetConfigurationOutput(ctx, talos_client.GetConfigurationOutputArgs{
		ClusterName:         pulumi.String("talos-cluster"),
		ClientConfiguration: clientConfigInput,
		Endpoints:           pulumi.StringArray{pulumi.String(vipIP)},
	})

	ctx.Export("talosconfig", clientConfig.TalosConfig())

	// Write Talosconfig to local file
	clientConfig.TalosConfig().ApplyT(func(tc string) (interface{}, error) {
		err := os.WriteFile("talosconfig", []byte(tc), 0600)
		return nil, err
	})

	// Define nodes (HA Control Plane + Scaled Workers)
	nodes := []NodeConfig{
		// Control Plane (3 Nodes)
		{Name: "k8s-controller1", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30},
		{Name: "k8s-controller2", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30},
		{Name: "k8s-controller3", Role: "control", Cores: 2, Memory: 4096, DiskSize: 30},
		// Workers (4 Nodes)
		{Name: "k8s-worker1", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125},
		{Name: "k8s-worker2", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125},
		{Name: "k8s-worker3", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125},
		// Worker 4 with GPU
		{
			Name: "k8s-worker4", Role: "worker", Cores: 4, Memory: 8192, DiskSize: 125,
			HasGPU: true, PcieIDs: []string{"0000:28:00.0"},
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
`

	basePatch := `cluster:
  network:
    cni:
      name: none
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

	// Export sanitized configs (used as base)
	// We will patch hostname per-node later
	ctx.Export("cp-config", cpConfig.MachineConfiguration())
	ctx.Export("worker-config", workerConfig.MachineConfiguration())

	var cpConfigApplies []pulumi.Resource
	var controllerIP pulumi.StringOutput

	for _, node := range nodes {
		// Create VM
		vmRes, err := createVM(ctx, provider, cfg.NodeName, node, talosImage)
		if err != nil {
			return err
		}

		// Capture IP from Agent
		// Note: This commonly fails on the first run if the VM is too slow to report an IP.
		// If it errors, run `pulumi refresh` and `pulumi up` again.
		nodeIP := vmRes.Ipv4Addresses.ApplyT(func(ips [][]string) (string, error) {
			for _, netInterface := range ips {
				for _, ip := range netInterface {
					if ip != "127.0.0.1" && !strings.Contains(ip, ":") {
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
	// Depends on Control Plane configs being applied
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
	kubeconfigRes.KubeconfigRaw.ApplyT(func(kc string) (interface{}, error) {
		err := os.WriteFile("kubeconfig", []byte(kc), 0600)
		return nil, err
	})

	ctx.Export("kubeconfig", kubeconfigRes.KubeconfigRaw)

	return nil
}

// patchTalosConfig removes HostnameConfig, machine.install, and sets the hostname.
func patchTalosConfig(rawConfig, hostname string) (string, error) {
	decoder := yaml.NewDecoder(strings.NewReader(rawConfig))
	var documents []map[string]interface{}

	for {
		var doc map[string]interface{}
		err := decoder.Decode(&doc)
		if err != nil {
			if err.Error() == "EOF" {
				// Clean exit on EOF
				break
			}
			return "", err
		}

		// Filter 1: Remove HostnameConfig
		if kind, ok := doc["kind"].(string); ok && kind == "HostnameConfig" {
			continue
		}

		// Filter 2: Remove machine.install in MachineConfig
		if _, hasMachine := doc["machine"]; hasMachine {
			if machineMap, ok := doc["machine"].(map[string]interface{}); ok {
				delete(machineMap, "install")

				// Patch 3: Set Hostname
				if _, hasNetwork := machineMap["network"]; !hasNetwork {
					machineMap["network"] = make(map[string]interface{})
				}
				if networkMap, ok := machineMap["network"].(map[string]interface{}); ok {
					networkMap["hostname"] = hostname
				}
			}
		}

		documents = append(documents, doc)
	}

	// Re-marshal
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

func createVM(ctx *pulumi.Context, provider *proxmoxve.Provider, nodeName string, config NodeConfig, talosImage *download.File) (*vm.VirtualMachine, error) {
	diskArgs := vm.VirtualMachineDiskArgs{
		Size:        pulumi.Int(config.DiskSize),
		Interface:   pulumi.String("scsi0"),
		Iothread:    pulumi.Bool(true),
		FileFormat:  pulumi.String("raw"),
		FileId:      talosImage.ID(),
		DatastoreId: pulumi.String("local-lvm"),
	}

	args := &vm.VirtualMachineArgs{
		NodeName: pulumi.String(nodeName),
		Name:     pulumi.String(config.Name),
		Agent:    vm.VirtualMachineAgentArgs{Enabled: pulumi.Bool(true)},
		Cpu: &vm.VirtualMachineCpuArgs{
			Cores: pulumi.Int(config.Cores),
			Type:  pulumi.String("host"),
		},
		Memory: &vm.VirtualMachineMemoryArgs{
			Dedicated: pulumi.Int(config.Memory),
		},
		Disks: &vm.VirtualMachineDiskArray{
			diskArgs,
		},
		BootOrders:   pulumi.StringArray{pulumi.String("scsi0")},
		ScsiHardware: pulumi.String("virtio-scsi-single"),
		OperatingSystem: vm.VirtualMachineOperatingSystemArgs{
			Type: pulumi.String("l26"), // Linux 2.6+
		},
		NetworkDevices: vm.VirtualMachineNetworkDeviceArray{
			vm.VirtualMachineNetworkDeviceArgs{
				Model:  pulumi.String("virtio"),
				Bridge: pulumi.String("vmbr0"),
			},
		},
		OnBoot: pulumi.Bool(true),
		// Talos specific optimizations
		Bios:    pulumi.String("ovmf"),
		Machine: pulumi.String("q35"),
	}

	if config.HasGPU {
		var hostpcis vm.VirtualMachineHostpciArray
		for i, pcieID := range config.PcieIDs {
			hostpcis = append(hostpcis, vm.VirtualMachineHostpciArgs{
				Device: pulumi.String("hostpci0"),
				Pcie:   pulumi.Bool(true),
				Id:     pulumi.String(pcieID),
				Xvga:   pulumi.Bool(i == 0),
			})
		}
		args.Hostpcis = hostpcis
	}

	return vm.NewVirtualMachine(ctx, config.Name, args, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{talosImage}))
}
