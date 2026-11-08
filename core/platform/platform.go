// Package platform builds the cluster itself: Proxmox VMs and Talos (the talos
// stack), then the components Argo CD cannot deploy for itself (the platform
// stack): Cilium, Gateway API, cert-manager and Argo CD.
package platform

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DeployPlatform installs Gateway API CRDs, Cilium with the shared Gateway and
// LB IP pool, cert-manager and Argo CD. It reads the kubeconfig that the talos
// stack writes, so the talos stack must be up first.
func DeployPlatform(ctx *pulumi.Context) error {
	kubeconfigBytes, err := os.ReadFile("kubeconfig")
	if err != nil {
		return fmt.Errorf("kubeconfig not found — run `just core talos up` first: %w", err)
	}

	k8sProvider, err := kubernetes.NewProvider(ctx, "k8s-provider", &kubernetes.ProviderArgs{
		Kubeconfig: pulumi.String(string(kubeconfigBytes)),
	})
	if err != nil {
		return err
	}

	gatewayCRDs, err := InstallGatewayAPICRDs(ctx, k8sProvider)
	if err != nil {
		return err
	}

	ciliumChart, err := InstallCilium(ctx, k8sProvider, gatewayCRDs)
	if err != nil {
		return err
	}

	if err := InstallGateway(ctx, k8sProvider, gatewayCRDs, ciliumChart); err != nil {
		return err
	}

	if err := TaintGPUNodes(ctx, k8sProvider); err != nil {
		return err
	}

	if err := InstallCertManager(ctx, k8sProvider); err != nil {
		return err
	}

	if err := InstallArgoCD(ctx, k8sProvider); err != nil {
		return err
	}

	return ConfigureCiliumIPPool(ctx, k8sProvider, ciliumChart, "(eth0|ens.*|enp.*)")
}
