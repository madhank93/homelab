package platform

import (
	"errors"
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

// DeployTalosCluster provisions Talos VMs on Proxmox and bootstraps Kubernetes.
// It also sets up the Proxmox host's certificate and Authentik login, downloads
// the base and GPU Talos images, creates the VMs in config.yml, applies per-node
// machine configs (hostname, static IP, route), bootstraps etcd on the first
// controller, and writes talosconfig and kubeconfig for the platform stack.
//
// Changing the Talos image affects only newly created VMs; upgrade existing
// nodes with `just talos-upgrade <ip>`.
func DeployTalosCluster(ctx *pulumi.Context) error {
	provider, err := NewProxmoxProvider(ctx)
	if err != nil {
		return err
	}
	if err := NewProxmoxCert(ctx, provider); err != nil {
		return err
	}
	if err := NewProxmoxSSO(ctx, provider); err != nil {
		return err
	}
	cl := config.C.Cluster
	talos := config.V.Talos
	pveNode := config.C.Proxmox.NodeName
	clusterEndpoint := "https://" + cl.VIP + ":6443"
	factoryURL := func(schematic string) string {
		return "https://factory.talos.dev/image/" + schematic + "/" + talos.Version + "/nocloud-amd64.raw.gz"
	}

	// Base extensions: iscsi-tools, util-linux-tools, qemu-guest-agent.
	baseTalosImage, err := DownloadImage(ctx, provider, "talos-base-image", pveNode,
		factoryURL(talos.BaseSchematic),
		"talos-nocloud-amd64-base.img",
		"gz",
	)
	if err != nil {
		return err
	}

	// GPU image: the base extensions plus nvidia-container-toolkit-production and
	// nvidia-open-gpu-kernel-modules-production.
	gpuTalosImage, err := DownloadImage(ctx, provider, "talos-gpu-image", pveNode,
		factoryURL(talos.GPUSchematic),
		"talos-nocloud-amd64-gpu.img",
		"gz",
	)
	if err != nil {
		return err
	}

	// The provider replaces the secrets (new cluster PKI) when TalosVersion
	// decreases, so a version rollback must never reach pulumi up.
	secrets, err := talos_machine.NewSecrets(ctx, "talos-secrets", &talos_machine.SecretsArgs{
		TalosVersion: pulumi.String(talos.Version),
	}, pulumi.Protect(true))
	if err != nil {
		return err
	}

	clientConfigInput := talos_client.GetConfigurationClientConfigurationArgs{
		CaCertificate:     secrets.ClientConfiguration.CaCertificate(),
		ClientCertificate: secrets.ClientConfiguration.ClientCertificate(),
		ClientKey:         secrets.ClientConfiguration.ClientKey(),
	}

	clientConfig := talos_client.GetConfigurationOutput(ctx, talos_client.GetConfigurationOutputArgs{
		ClusterName:         pulumi.String(cl.Name),
		ClientConfiguration: clientConfigInput,
		Endpoints:           pulumi.StringArray{pulumi.String(cl.VIP)},
	})

	ctx.Export("talosconfig", clientConfig.TalosConfig())

	clientConfig.TalosConfig().ApplyT(func(tc string) (any, error) {
		err := os.WriteFile("talosconfig", []byte(tc), 0o600)
		return nil, err
	})

	// Control planes share the API VIP. The API server also trusts ID tokens from Authentik's "kubernetes" client,
	// so Headlamp logins act as oidc:<username> in oidc:<group>s. A down
	// Authentik only blocks new OIDC logins; certificate auth is unaffected.
	cpPatch := fmt.Sprintf(`machine:
  network:
    interfaces:
      - deviceSelector:
          physical: true
        dhcp: true
        vip:
          ip: %s
cluster:
  apiServer:
    extraArgs:
      oidc-issuer-url: https://%s/application/o/kubernetes/
      oidc-client-id: kubernetes
      oidc-username-claim: preferred_username
      oidc-username-prefix: "oidc:"
      oidc-groups-claim: groups
      oidc-groups-prefix: "oidc:"
`, cl.VIP, config.C.Host("auth"))

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
    # Images and Longhorn replicas share /var, so the 15%% imagefs default
    # evicts Longhorn itself before it can free space. Setting one threshold
    # zeroes the rest, hence the full list.
    extraConfig:
      evictionHard:
        memory.available: 100Mi
        nodefs.available: 10%%
        nodefs.inodesFree: 5%%
        imagefs.available: 10%%
        imagefs.inodesFree: 5%%
  kernel:
    modules:
      - name: nbd
      - name: iscsi_tcp
      - name: iscsi_generic
      - name: configfs
`, cl.Subnet)

	// Applied on top of workerPatch; Talos appends the module list. The
	// nvidia-container-toolkit extension configures containerd itself.
	// The kubelet adds the taint when the node first registers; TaintGPUNodes in
	// the platform stack keeps it on the existing node.
	gpuWorkerPatch := `machine:
  kubelet:
    extraConfig:
      registerWithTaints:
        - {key: dedicated, value: ai, effect: NoSchedule}
  kernel:
    modules:
      - name: nvidia
      - name: nvidia_uvm
      - name: nvidia_drm
      - name: nvidia_modeset
`

	// install.disk is required: these nodes boot with talos.platform=metal, and the
	// metal installer refuses to run an upgrade without it. /dev/sda is each VM's
	// only disk (scsi0, sized per node in config.yml). install.image stays unset:
	// `just talos-upgrade` passes the Image Factory installer explicitly.
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

	// KubernetesVersion pins the kubelet and control-plane images;
	// left unset, the provider's built-in default would roll nodes back on apply.
	cpConfig := talos_machine.GetConfigurationOutput(ctx, talos_machine.GetConfigurationOutputArgs{
		ClusterName:       pulumi.String(cl.Name),
		MachineType:       pulumi.String("controlplane"),
		ClusterEndpoint:   pulumi.String(clusterEndpoint),
		MachineSecrets:    secrets.MachineSecrets,
		TalosVersion:      pulumi.String(talos.Version),
		KubernetesVersion: pulumi.String(config.V.Kubernetes.Version),
		ConfigPatches:     pulumi.StringArray{pulumi.String(basePatch), pulumi.String(cpPatch)},
	})

	workerConfig := talos_machine.GetConfigurationOutput(ctx, talos_machine.GetConfigurationOutputArgs{
		ClusterName:       pulumi.String(cl.Name),
		MachineType:       pulumi.String("worker"),
		ClusterEndpoint:   pulumi.String(clusterEndpoint),
		MachineSecrets:    secrets.MachineSecrets,
		TalosVersion:      pulumi.String(talos.Version),
		KubernetesVersion: pulumi.String(config.V.Kubernetes.Version),
		ConfigPatches:     pulumi.StringArray{pulumi.String(basePatch), pulumi.String(workerPatch)},
	})

	gpuConfig := talos_machine.GetConfigurationOutput(ctx, talos_machine.GetConfigurationOutputArgs{
		ClusterName:       pulumi.String(cl.Name),
		MachineType:       pulumi.String("worker"),
		ClusterEndpoint:   pulumi.String(clusterEndpoint),
		MachineSecrets:    secrets.MachineSecrets,
		TalosVersion:      pulumi.String(talos.Version),
		KubernetesVersion: pulumi.String(config.V.Kubernetes.Version),
		ConfigPatches:     pulumi.StringArray{pulumi.String(basePatch), pulumi.String(workerPatch), pulumi.String(gpuWorkerPatch)},
	})

	ctx.Export("cp-config", cpConfig.MachineConfiguration())
	ctx.Export("worker-config", workerConfig.MachineConfiguration())

	clientConfigArgs := talos_machine.ClientConfigurationArgs{
		CaCertificate:     secrets.ClientConfiguration.CaCertificate(),
		ClientCertificate: secrets.ClientConfiguration.ClientCertificate(),
		ClientKey:         secrets.ClientConfiguration.ClientKey(),
	}

	var cpConfigApplies []pulumi.Resource
	var controllerIP pulumi.String

	for _, node := range cl.Nodes {
		var imageID pulumi.IDOutput
		if node.HasGPU() {
			imageID = gpuTalosImage.ID()
		} else {
			imageID = baseTalosImage.ID()
		}

		vmRes, err := NewProxmoxVM(ctx, provider, pveNode, node, imageID)
		if err != nil {
			return err
		}

		nodeIP := pulumi.String(node.IP)

		ctx.Export(node.Name+"-ip", nodeIP)

		if controllerIP == "" && node.Role == "control" {
			controllerIP = nodeIP
		}

		var baseConfig pulumi.StringOutput
		switch {
		case node.Role == "control":
			baseConfig = cpConfig.MachineConfiguration()
		case node.HasGPU():
			baseConfig = gpuConfig.MachineConfiguration()
		default:
			baseConfig = workerConfig.MachineConfiguration()
		}

		finalConfig := baseConfig.ApplyT(func(c string) (string, error) {
			return patchTalosConfig(c, node.Name, node.IP, cl)
		}).(pulumi.StringOutput)

		// Reach the node at the address its QEMU agent reports, falling back to the
		// static IP so a stopped VM does not block preview for the other nodes.
		dynamicEndpoint := vmRes.Ipv4Addresses.ApplyT(func(ips [][]string) (string, error) {
			return firstRoutableIPv4(ips, node.IP), nil
		}).(pulumi.StringOutput)

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

	// etcd is bootstrapped exactly once, on the first control-plane node in config.yml.
	bootstrap, err := talos_machine.NewBootstrap(ctx, "bootstrap", &talos_machine.BootstrapArgs{
		ClientConfiguration: clientConfigArgs,
		Node:                controllerIP,
	}, pulumi.DependsOn(cpConfigApplies))
	if err != nil {
		return err
	}

	kubeconfigRes, err := talos_cluster.NewKubeconfig(ctx, "kubeconfig", &talos_cluster.KubeconfigArgs{
		ClientConfiguration: talos_cluster.KubeconfigClientConfigurationArgs(clientConfigArgs),
		Node:                pulumi.String(cl.VIP),
	}, pulumi.DependsOn([]pulumi.Resource{bootstrap}))
	if err != nil {
		return err
	}

	// DeployPlatform reads this file.
	kubeconfigRes.KubeconfigRaw.ApplyT(func(kc string) (any, error) {
		err := os.WriteFile("kubeconfig", []byte(kc), 0o600)
		return nil, err
	})

	ctx.Export("kubeconfig", kubeconfigRes.KubeconfigRaw)

	return nil
}

// firstRoutableIPv4 returns the first QEMU-agent address that is not loopback,
// link-local or IPv6, or fallback when there is none (VM stopped or agent not
// ready).
func firstRoutableIPv4(ips [][]string, fallback string) string {
	for _, net := range ips {
		for _, ip := range net {
			if ip != "127.0.0.1" && ip != "" && !strings.HasPrefix(ip, "169.254.") && !strings.Contains(ip, ":") {
				return ip
			}
		}
	}
	return fallback
}

// patchTalosConfig applies per-node patches to a raw Talos machine config YAML.
//
// It drops the generated HostnameConfig document, which would conflict with the
// explicit hostname, and applies [patchMachineNetwork] to the machine document.
func patchTalosConfig(rawConfig, hostname, ip string, cl config.ClusterConfig) (string, error) {
	_, bits, ok := strings.Cut(cl.Subnet, "/")
	if !ok {
		return "", fmt.Errorf("cluster subnet %q is not in CIDR form", cl.Subnet)
	}
	prefix := "/" + bits
	reader := strings.NewReader(rawConfig)
	decoder := yaml.NewDecoder(reader)
	var documents []map[string]any

	for {
		var doc map[string]any
		err := decoder.Decode(&doc)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}

		if kind, ok := doc["kind"].(string); ok && kind == "HostnameConfig" {
			continue
		}

		if machineMap, ok := doc["machine"].(map[string]any); ok {
			patchMachineNetwork(machineMap, hostname, ip+prefix, cl)
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
	if err := encoder.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// patchMachineNetwork sets the hostname and nameservers on a machine document
// and turns its first interface from DHCP into a static address with a default
// route via cl.Router.
func patchMachineNetwork(machineMap map[string]any, hostname, address string, cl config.ClusterConfig) {
	if _, hasNetwork := machineMap["network"]; !hasNetwork {
		machineMap["network"] = make(map[string]any)
	}
	networkMap, ok := machineMap["network"].(map[string]any)
	if !ok {
		return
	}
	networkMap["hostname"] = hostname
	if interfaces, ok := networkMap["interfaces"].([]any); ok && len(interfaces) > 0 {
		if iface, ok := interfaces[0].(map[string]any); ok {
			iface["dhcp"] = false
			iface["addresses"] = []string{address}
			iface["routes"] = []any{
				map[string]any{
					"gateway": cl.Router,
				},
			}
		}
	}
	networkMap["nameservers"] = cl.Nameservers
}
