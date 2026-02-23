package seccomp

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/k8s"
)

func NewSealedSecretsChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Install CRD from v0.35.0 release (matches controller version)
	cdk8s.NewInclude(chart, jsii.String("sealed-secrets-crd"), &cdk8s.IncludeProps{
		Url: jsii.String("https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.35.0/controller.yaml"),
	})

	// Install controller via Helm with custom image and RBAC disabled
	cdk8s.NewHelm(chart, jsii.String("sealed-secrets"), &cdk8s.HelmProps{
		Chart:       jsii.String("sealed-secrets"),
		Namespace:   jsii.String("kube-system"),
		Repo:        jsii.String("https://bitnami-labs.github.io/sealed-secrets"),
		ReleaseName: jsii.String("sealed-secrets-controller"),
		Version:     jsii.String("2.18.1"),
		Values: &map[string]any{
			"fullnameOverride": "sealed-secrets-controller",
			"image": map[string]any{
				"registry":   "",
				"repository": "ghcr.io/bitnami-labs/sealed-secrets-controller",
				"tag":        "0.35.0",
			},
			"rbac": map[string]any{
				"create": true, // Use default chart RBAC
			},
			"serviceAccount": map[string]any{
				"create": true,
				"name":   "sealed-secrets-controller",
			},
			"crd": map[string]any{
				"create": false, // Installed via Include above
				"keep":   true,
			},
			"resources": map[string]any{
				"limits": map[string]any{
					"cpu":    "200m",
					"memory": "256Mi",
				},
				"requests": map[string]any{
					"cpu":    "50m",
					"memory": "64Mi",
				},
			},
			"securityContext": map[string]any{
				"runAsUser":              1001,
				"runAsNonRoot":           true,
				"readOnlyRootFilesystem": true,
			},
		},
	})

	// Add RBAC patch for missing 'list' permission on secrets
	k8s.NewKubeClusterRole(chart, jsii.String("secrets-unsealer-patch"), &k8s.KubeClusterRoleProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String("secrets-unsealer-patch"),
			Labels: &map[string]*string{
				"app.kubernetes.io/name": jsii.String("sealed-secrets"),
			},
		},
		Rules: &[]*k8s.PolicyRule{
			{
				ApiGroups: &[]*string{jsii.String("")},
				Resources: &[]*string{jsii.String("secrets")},
				Verbs: &[]*string{
					jsii.String("get"),
					jsii.String("list"),
					jsii.String("watch"),
					jsii.String("create"),
					jsii.String("update"),
					jsii.String("delete"),
				},
			},
		},
	})

	// Bind patch role to controller service account
	k8s.NewKubeClusterRoleBinding(chart, jsii.String("sealed-secrets-controller-patch"), &k8s.KubeClusterRoleBindingProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String("sealed-secrets-controller-patch"),
			Labels: &map[string]*string{
				"app.kubernetes.io/name": jsii.String("sealed-secrets"),
			},
		},
		RoleRef: &k8s.RoleRef{
			ApiGroup: jsii.String("rbac.authorization.k8s.io"),
			Kind:     jsii.String("ClusterRole"),
			Name:     jsii.String("secrets-unsealer-patch"),
		},
		Subjects: &[]*k8s.Subject{
			{
				Kind:      jsii.String("ServiceAccount"),
				Name:      jsii.String("sealed-secrets-controller"),
				Namespace: jsii.String("kube-system"),
			},
		},
	})

	return chart
}
