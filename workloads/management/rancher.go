package management

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

func NewRancherChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Create namespace
	cdk8s.NewApiObject(chart, jsii.String("rancher-namespace"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Namespace"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String(namespace),
		},
	})

	// SecretProviderClass — Pattern B (secretObjects sync).
	// Fetches BOOTSTRAP_PASSWORD from OpenBao and syncs it into k8s Secret "rancher-bootstrap".
	// Rancher Helm chart references existingBootstrapPassword: "rancher-bootstrap".
	cdk8s.NewApiObject(chart, jsii.String("rancher-spc"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets-store.csi.x-k8s.io/v1"),
		Kind:       jsii.String("SecretProviderClass"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("rancher-secrets"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"provider": "openbao",
		"parameters": map[string]any{
			"vaultAddress": "http://openbao.openbao.svc.cluster.local:8200",
			"roleName":     "rancher",
			"objects": `- objectName: "BOOTSTRAP_PASSWORD"
  secretPath: "secret/data/rancher"
  secretKey: "BOOTSTRAP_PASSWORD"`,
		},
		"secretObjects": []map[string]any{
			{
				"secretName": "rancher-bootstrap",
				"type":       "Opaque",
				"data": []map[string]any{
					{
						"objectName": "BOOTSTRAP_PASSWORD",
						"key":        "BOOTSTRAP_PASSWORD",
					},
				},
			},
		},
	}))

	// ServiceAccount for the secret-sync pod.
	// The OpenBao K8s auth role "rancher" is bound to this SA (see scripts/openbao-setup.sh).
	k8s.NewKubeServiceAccount(chart, jsii.String("rancher-secret-sync-sa"), &k8s.KubeServiceAccountProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("secret-sync"),
			Namespace: jsii.String(namespace),
		},
	})

	// Secret-sync Deployment — mounts the CSI volume to trigger secretObjects sync.
	// Rancher's Helm chart does not support extraVolumes on its pods, so this dedicated
	// pod triggers creation of the rancher-bootstrap Secret.
	replicas := float64(1)
	k8s.NewKubeDeployment(chart, jsii.String("rancher-secret-sync"), &k8s.KubeDeploymentProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("secret-sync"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.DeploymentSpec{
			Replicas: &replicas,
			Selector: &k8s.LabelSelector{
				MatchLabels: &map[string]*string{"app": jsii.String("secret-sync")},
			},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{
					Labels: &map[string]*string{"app": jsii.String("secret-sync")},
				},
				Spec: &k8s.PodSpec{
					ServiceAccountName: jsii.String("secret-sync"),
					Volumes: &[]*k8s.Volume{
						{
							Name: jsii.String("openbao-secrets"),
							Csi: &k8s.CsiVolumeSource{
								Driver:   jsii.String("secrets-store.csi.k8s.io"),
								ReadOnly: jsii.Bool(true),
								VolumeAttributes: &map[string]*string{
									"secretProviderClass": jsii.String("rancher-secrets"),
								},
							},
						},
					},
					Containers: &[]*k8s.Container{
						{
							Name:  jsii.String("pause"),
							Image: jsii.String("registry.k8s.io/pause:3.10"),
							VolumeMounts: &[]*k8s.VolumeMount{
								{
									Name:      jsii.String("openbao-secrets"),
									MountPath: jsii.String("/mnt/secrets"),
									ReadOnly:  jsii.Bool(true),
								},
							},
						},
					},
				},
			},
		},
	})

	values := map[string]any{
		"agentTLSMode": "system-store",
		"auditLog": map[string]any{
			"level":       0,
			"destination": "sidecar",
		},
		// Ingress disabled — traffic routed via Gateway API HTTPRoute below
		// Rancher's built-in ingress required an Nginx controller; removed in favour of homelab-gateway
		"ingress": map[string]any{
			"enabled": false,
		},
		"service": map[string]any{
			"type":        "ClusterIP",
			"disableHttp": false,
		},
		"hostname":                   "rancher.madhan.app", // Updated to real domain
		"bootstrapPassword":          "",                   // Not used when secret exists
		"existingBootstrapPassword":  "rancher-bootstrap",  // Secret synced by SecretProviderClass + secret-sync pod
		"bootstrapPasswordSecretKey": "BOOTSTRAP_PASSWORD",
		"replicas":                   3,
		"resources": map[string]any{
			"limits": map[string]any{
				"memory": "2Gi",
				"cpu":    "1000m",
			},
			"requests": map[string]any{
				"memory": "1Gi",
				"cpu":    "500m",
			},
		},
		"antiAffinity": "preferred",
	}

	cdk8s.NewHelm(chart, jsii.String("rancher-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("rancher"),
		Repo:        jsii.String("https://releases.rancher.com/server-charts/stable"),
		ReleaseName: jsii.String("rancher"),
		Namespace:   jsii.String(namespace),
		Version:     jsii.String("2.13.2"),
		HelmFlags:   &[]*string{jsii.String("--kube-version"), jsii.String("1.30.0")},
		Values:      &values,
	})

	// Gateway API HTTPRoute — routes rancher.madhan.app → rancher:80
	cdk8s.NewApiObject(chart, jsii.String("rancher-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("rancher"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"rancher.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "rancher", "port": 80, "weight": 1},
				},
			},
		},
	}))

	return chart
}
