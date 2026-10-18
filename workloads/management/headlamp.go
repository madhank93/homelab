package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/headlamp"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewHeadlampChart deploys Headlamp, a Kubernetes web UI, via the official Helm chart.
func NewHeadlampChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	headlamp.NewHeadlamp(chart, jsii.String("headlamp-release"), &headlamp.HeadlampProps{
		ReleaseName: jsii.String("headlamp"),
		Namespace:   jsii.String(namespace),
		Values: &headlamp.HeadlampValues{
			AdditionalValues: &map[string]interface{}{
				"resources": map[string]interface{}{
					"limits":   map[string]interface{}{"cpu": "500m", "memory": "512Mi"},
					"requests": map[string]interface{}{"cpu": "100m", "memory": "128Mi"},
				},
				// Exposed through the Gateway instead.
				"ingress": map[string]interface{}{"enabled": false},
			},
		},
	})

	// Login identity for the UI. No static token Secret: mint a short-lived one
	// with `just headlamp-token` (kubectl create token).
	k8s.NewKubeServiceAccount(chart, jsii.String("headlamp-admin-sa"), &k8s.KubeServiceAccountProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("headlamp-admin"),
			Namespace: jsii.String(namespace),
		},
	})

	k8s.NewKubeClusterRoleBinding(chart, jsii.String("headlamp-admin-crb"), &k8s.KubeClusterRoleBindingProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String("headlamp-admin"),
		},
		RoleRef: &k8s.RoleRef{
			ApiGroup: jsii.String("rbac.authorization.k8s.io"),
			Kind:     jsii.String("ClusterRole"),
			Name:     jsii.String("cluster-admin"),
		},
		Subjects: &[]*k8s.Subject{
			{
				Kind:      jsii.String("ServiceAccount"),
				Name:      jsii.String("headlamp-admin"),
				Namespace: jsii.String(namespace),
			},
		},
	})

	kit.Route(chart, "headlamp", "headlamp", "headlamp", 80)

	return chart
}
