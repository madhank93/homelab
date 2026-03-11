package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/headlamp"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

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
				// Ingress disabled — traffic routed via Gateway API HTTPRoute below
				"ingress": map[string]interface{}{"enabled": false},
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

	// ServiceAccount for read-only public access (no logs, no exec, no secrets)
	k8s.NewKubeServiceAccount(chart, jsii.String("headlamp-readonly-sa"), &k8s.KubeServiceAccountProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("headlamp-readonly"),
			Namespace: jsii.String(namespace),
		},
	})

	// Custom ClusterRole — mirrors built-in "view" but omits pods/log.
	// Kubernetes RBAC has no deny rules, so pods/log must be excluded at the ClusterRole level.
	k8s.NewKubeClusterRole(chart, jsii.String("headlamp-readonly-cr"), &k8s.KubeClusterRoleProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String("headlamp-readonly"),
		},
		Rules: &[]*k8s.PolicyRule{
			{
				ApiGroups: &[]*string{jsii.String("")},
				Resources: &[]*string{
					jsii.String("configmaps"),
					jsii.String("endpoints"),
					jsii.String("events"),
					jsii.String("limitranges"),
					jsii.String("namespaces"),
					jsii.String("namespaces/status"),
					jsii.String("nodes"),
					jsii.String("persistentvolumeclaims"),
					jsii.String("persistentvolumeclaims/status"),
					jsii.String("pods"),
					jsii.String("pods/status"),
					// pods/log intentionally omitted
					jsii.String("replicationcontrollers"),
					jsii.String("replicationcontrollers/scale"),
					jsii.String("replicationcontrollers/status"),
					jsii.String("resourcequotas"),
					jsii.String("resourcequotas/status"),
					jsii.String("serviceaccounts"),
					jsii.String("services"),
					jsii.String("services/status"),
				},
				Verbs: &[]*string{jsii.String("get"), jsii.String("list"), jsii.String("watch")},
			},
			{
				ApiGroups: &[]*string{jsii.String("apps")},
				Resources: &[]*string{
					jsii.String("controllerrevisions"),
					jsii.String("daemonsets"),
					jsii.String("daemonsets/status"),
					jsii.String("deployments"),
					jsii.String("deployments/scale"),
					jsii.String("deployments/status"),
					jsii.String("replicasets"),
					jsii.String("replicasets/scale"),
					jsii.String("replicasets/status"),
					jsii.String("statefulsets"),
					jsii.String("statefulsets/scale"),
					jsii.String("statefulsets/status"),
				},
				Verbs: &[]*string{jsii.String("get"), jsii.String("list"), jsii.String("watch")},
			},
			{
				ApiGroups: &[]*string{jsii.String("autoscaling")},
				Resources: &[]*string{
					jsii.String("horizontalpodautoscalers"),
					jsii.String("horizontalpodautoscalers/status"),
				},
				Verbs: &[]*string{jsii.String("get"), jsii.String("list"), jsii.String("watch")},
			},
			{
				ApiGroups: &[]*string{jsii.String("batch")},
				Resources: &[]*string{
					jsii.String("cronjobs"),
					jsii.String("cronjobs/status"),
					jsii.String("jobs"),
					jsii.String("jobs/status"),
				},
				Verbs: &[]*string{jsii.String("get"), jsii.String("list"), jsii.String("watch")},
			},
			{
				ApiGroups: &[]*string{jsii.String("policy")},
				Resources: &[]*string{
					jsii.String("poddisruptionbudgets"),
					jsii.String("poddisruptionbudgets/status"),
				},
				Verbs: &[]*string{jsii.String("get"), jsii.String("list"), jsii.String("watch")},
			},
			{
				ApiGroups: &[]*string{jsii.String("networking.k8s.io")},
				Resources: &[]*string{
					jsii.String("ingresses"),
					jsii.String("ingresses/status"),
					jsii.String("networkpolicies"),
				},
				Verbs: &[]*string{jsii.String("get"), jsii.String("list"), jsii.String("watch")},
			},
			{
				ApiGroups: &[]*string{jsii.String("storage.k8s.io")},
				Resources: &[]*string{
					jsii.String("storageclasses"),
					jsii.String("volumeattachments"),
					jsii.String("volumeattachments/status"),
				},
				Verbs: &[]*string{jsii.String("get"), jsii.String("list"), jsii.String("watch")},
			},
		},
	})

	// ClusterRoleBinding — binds headlamp-readonly SA to the custom ClusterRole above
	k8s.NewKubeClusterRoleBinding(chart, jsii.String("headlamp-readonly-crb"), &k8s.KubeClusterRoleBindingProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String("headlamp-readonly"),
		},
		RoleRef: &k8s.RoleRef{
			ApiGroup: jsii.String("rbac.authorization.k8s.io"),
			Kind:     jsii.String("ClusterRole"),
			Name:     jsii.String("headlamp-readonly"),
		},
		Subjects: &[]*k8s.Subject{
			{
				Kind:      jsii.String("ServiceAccount"),
				Name:      jsii.String("headlamp-readonly"),
				Namespace: jsii.String(namespace),
			},
		},
	})

	// Long-lived read-only token — share with public users for headlamp.proxy.madhan.app
	// Retrieve with: kubectl get secret headlamp-readonly-token -n headlamp -o jsonpath='{.data.token}' | base64 -d
	k8s.NewKubeSecret(chart, jsii.String("headlamp-readonly-token"), &k8s.KubeSecretProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("headlamp-readonly-token"),
			Namespace: jsii.String(namespace),
			Annotations: &map[string]*string{
				"kubernetes.io/service-account.name": jsii.String("headlamp-readonly"),
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
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"headlamp.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "headlamp", "port": 80, "weight": 1},
				},
			},
		},
	}))

	return chart
}
