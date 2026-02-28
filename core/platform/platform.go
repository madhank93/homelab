package platform

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DeployPlatform installs cluster-level platform tools: Cilium, Gateway API,
// cert-manager, and ArgoCD. It reads the kubeconfig written by the talos stack.
// Run `just pulumi talos up` before running `just pulumi platform up`.
func DeployPlatform(ctx *pulumi.Context) error {
	kubeconfigBytes, err := os.ReadFile("kubeconfig")
	if err != nil {
		return fmt.Errorf("kubeconfig not found — run `just pulumi talos up` first: %w", err)
	}

	k8sProvider, err := kubernetes.NewProvider(ctx, "k8s-provider", &kubernetes.ProviderArgs{
		Kubeconfig: pulumi.String(string(kubeconfigBytes)),
	})
	if err != nil {
		return err
	}

	if err := InstallCilium(ctx, k8sProvider); err != nil {
		return err
	}

	if err := InstallGateway(ctx, k8sProvider); err != nil {
		return err
	}

	if err := InstallCertManager(ctx, k8sProvider); err != nil {
		return err
	}

	if err := InstallArgoCD(ctx, k8sProvider); err != nil {
		return err
	}

	// k8s-controller1 has a static IP — used to determine the LAN subnet for the IP pool
	if err := ConfigureCiliumIPPool(ctx, k8sProvider, pulumi.String("192.168.1.211"), "(eth0|ens.*|enp.*)"); err != nil {
		return err
	}

	if err := PatchCiliumRBAC(ctx, k8sProvider); err != nil {
		return err
	}

	return nil
}
