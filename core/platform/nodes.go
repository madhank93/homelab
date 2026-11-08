package platform

import (
	"github.com/madhank93/homelab/config"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// TaintGPUNodes keeps dedicated=ai:NoSchedule on every GPU node. NodeRestriction
// stops a kubelet changing its own taints, so Talos's machine.nodeTaints cannot
// apply it; the kubelet's registerWithTaints covers only a node's first registration.
func TaintGPUNodes(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {
	for _, node := range config.C.Cluster.Nodes {
		if !node.HasGPU() {
			continue
		}
		_, err := corev1.NewNodePatch(ctx, node.Name+"-gpu-taint", &corev1.NodePatchArgs{
			Metadata: &metav1.ObjectMetaPatchArgs{
				Name: pulumi.String(node.Name),
				// spec.taints is an atomic list that the node controller also writes
				// (cordon, NotReady), so take it back on every up.
				Annotations: pulumi.StringMap{"pulumi.com/patchForce": pulumi.String("true")},
			},
			Spec: &corev1.NodeSpecPatchArgs{
				Taints: corev1.TaintPatchArray{&corev1.TaintPatchArgs{
					Key:    pulumi.String("dedicated"),
					Value:  pulumi.String("ai"),
					Effect: pulumi.String("NoSchedule"),
				}},
			},
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return err
		}
	}
	return nil
}
