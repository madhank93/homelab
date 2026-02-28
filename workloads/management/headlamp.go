package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

func NewHeadlampChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	cdk8s.NewHelm(chart, jsii.String("headlamp-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("headlamp"),
		Repo:        jsii.String("https://kubernetes-sigs.github.io/headlamp/"),
		Version:     jsii.String("0.40.0"), // Pinned (released 2026-02-05)
		ReleaseName: jsii.String("headlamp"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
			// Ingress disabled — traffic routed via Gateway API HTTPRoute below
			"ingress": map[string]any{
				"enabled": false,
			},
		},
	})

	// ServiceAccount for Headlamp admin access
	k8s.NewKubeServiceAccount(chart, jsii.String("headlamp-admin-sa"), &k8s.KubeServiceAccountProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("headlamp-admin"),
			Namespace: jsii.String(namespace),
		},
	})

	// ClusterRoleBinding — grants headlamp-admin cluster-admin access
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

	// Long-lived token Secret — Kubernetes auto-populates .data.token
	// Retrieve with: kubectl get secret headlamp-admin-token -n headlamp -o jsonpath='{.data.token}' | base64 -d
	k8s.NewKubeSecret(chart, jsii.String("headlamp-admin-token"), &k8s.KubeSecretProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("headlamp-admin-token"),
			Namespace: jsii.String(namespace),
			Annotations: &map[string]*string{
				"kubernetes.io/service-account.name": jsii.String("headlamp-admin"),
			},
		},
		Type: jsii.String("kubernetes.io/service-account-token"),
	})

	// Gateway API HTTPRoute — routes headlamp.madhan.app → headlamp:80
	cdk8s.NewApiObject(chart, jsii.String("headlamp-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("headlamp"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"headlamp.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"name": "headlamp", "port": 80},
				},
			},
		},
	}))

	return chart
}
