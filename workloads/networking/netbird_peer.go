package networking

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

func NewNetbirdPeerChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Privileged namespace — netbird peer needs NET_ADMIN and SYS_MODULE for WireGuard.
	k8s.NewKubeNamespace(chart, jsii.String("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
			Labels: &map[string]*string{
				"pod-security.kubernetes.io/enforce": jsii.String("privileged"),
			},
		},
	})

	// InfisicalSecret: syncs NETBIRD_SETUP_KEY from /netbird path → secret netbird-setup-key.
	// Prerequisite: add NETBIRD_SETUP_KEY at path /netbird in Infisical before ArgoCD sync.
	infisicalSpec := map[string]any{
		"hostAPI":        "http://infisical-infisical-standalone-infisical.infisical.svc.cluster.local:8080",
		"resyncInterval": 60,
		"authentication": map[string]any{
			"serviceToken": map[string]any{
				"serviceTokenSecretReference": map[string]any{
					"secretName":      "infisical-service-token",
					"secretNamespace": "infisical",
				},
				"secretsScope": map[string]any{
					"projectSlug": "homelab-prod",
					"envSlug":     "prod",
					"secretsPath": "/netbird",
				},
			},
		},
		"managedSecretReference": map[string]any{
			"secretName":      "netbird-setup-key",
			"secretNamespace": namespace,
			"creationPolicy":  "Owner",
		},
	}

	cdk8s.NewApiObject(chart, jsii.String("netbird-infisical-secret"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets.infisical.com/v1alpha1"),
		Kind:       jsii.String("InfisicalSecret"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("netbird-secrets"),
			Namespace: jsii.String(namespace),
			// ServerSideApply=false: Infisical CRD schema omits projectSlug from
			// serviceToken.secretsScope, causing SSA schema validation to fail.
			Annotations: &map[string]*string{
				"argocd.argoproj.io/sync-options": jsii.String("ServerSideApply=false"),
			},
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), infisicalSpec))

	// Deployment: k8s-routing-peer — advertises 192.168.1.0/24 into the NetBird mesh.
	// hostNetwork: true so WireGuard can manipulate host routing table.
	// dnsPolicy: ClusterFirstWithHostNet to retain in-cluster DNS resolution.
	replicas := float64(1)
	k8s.NewKubeDeployment(chart, jsii.String("netbird-peer"), &k8s.KubeDeploymentProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("netbird-peer"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.DeploymentSpec{
			Replicas: &replicas,
			Selector: &k8s.LabelSelector{
				MatchLabels: &map[string]*string{"app": jsii.String("netbird-peer")},
			},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{
					Labels: &map[string]*string{"app": jsii.String("netbird-peer")},
				},
				Spec: &k8s.PodSpec{
					HostNetwork: jsii.Bool(true),
					DnsPolicy:   jsii.String("ClusterFirstWithHostNet"),
					Containers: &[]*k8s.Container{
						{
							Name:  jsii.String("netbird"),
							Image: jsii.String("netbirdio/netbird:latest"),
							Command: &[]*string{
								jsii.String("netbird"),
								jsii.String("up"),
								jsii.String("--advertise-routes=192.168.1.0/24"),
								jsii.String("--hostname=k8s-routing-peer"),
							},
							Env: &[]*k8s.EnvVar{
								{
									Name: jsii.String("NB_SETUP_KEY"),
									ValueFrom: &k8s.EnvVarSource{
										SecretKeyRef: &k8s.SecretKeySelector{
											Name: jsii.String("netbird-setup-key"),
											Key:  jsii.String("NETBIRD_SETUP_KEY"),
										},
									},
								},
								{
									Name:  jsii.String("NB_MANAGEMENT_URL"),
									Value: jsii.String("https://netbird.madhan.app"),
								},
							},
							SecurityContext: &k8s.SecurityContext{
								Capabilities: &k8s.Capabilities{
									Add: &[]*string{
										jsii.String("NET_ADMIN"),
										jsii.String("SYS_MODULE"),
									},
								},
							},
						},
					},
				},
			},
		},
	})

	return chart
}
