package registry

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

func NewHarborChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Create namespace
	cdk8s.NewApiObject(chart, jsii.String("harbor-namespace"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Namespace"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String(namespace),
		},
	})

	// SecretProviderClass — Pattern B (secretObjects sync).
	// Fetches HARBOR_ADMIN_PASSWORD from OpenBao and syncs it into k8s Secret "harbor-admin".
	// Harbor Helm chart references existingSecret: "harbor-admin".
	cdk8s.NewApiObject(chart, jsii.String("harbor-spc"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets-store.csi.x-k8s.io/v1"),
		Kind:       jsii.String("SecretProviderClass"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("harbor-secrets"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"provider": "openbao",
		"parameters": map[string]any{
			"vaultAddress": "http://openbao.openbao.svc.cluster.local:8200",
			"roleName":     "harbor",
			"objects": `- objectName: "HARBOR_ADMIN_PASSWORD"
  secretPath: "secret/data/harbor"
  secretKey: "HARBOR_ADMIN_PASSWORD"`,
		},
		// secretObjects: CSI driver creates/maintains this k8s Secret when a pod mounts this SPC.
		"secretObjects": []map[string]any{
			{
				"secretName": "harbor-admin",
				"type":       "Opaque",
				"data": []map[string]any{
					{
						"objectName": "HARBOR_ADMIN_PASSWORD",
						"key":        "HARBOR_ADMIN_PASSWORD",
					},
				},
			},
		},
	}))

	// ServiceAccount for the secret-sync pod.
	// The OpenBao K8s auth role "harbor" is bound to this SA (see scripts/openbao-setup.sh).
	k8s.NewKubeServiceAccount(chart, jsii.String("harbor-secret-sync-sa"), &k8s.KubeServiceAccountProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("secret-sync"),
			Namespace: jsii.String(namespace),
		},
	})

	// Secret-sync Deployment — mounts the CSI volume to trigger secretObjects sync.
	// Harbor's Helm chart does not support extraVolumes on its component pods,
	// so this dedicated pod is the trigger for creating the harbor-admin Secret.
	// The pause container has minimal resource usage.
	replicas := float64(1)
	k8s.NewKubeDeployment(chart, jsii.String("harbor-secret-sync"), &k8s.KubeDeploymentProps{
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
									"secretProviderClass": jsii.String("harbor-secrets"),
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
		"expose": map[string]any{
			// Use clusterIP; TLS terminated at homelab-gateway (wildcard-madhan-app-tls)
			"type": "clusterIP",
			"tls": map[string]any{
				"enabled": false,
			},
		},
		"externalURL": "https://harbor.madhan.app",
		"persistence": map[string]any{
			"enabled": true,
			"persistentVolumeClaim": map[string]any{
				"registry": map[string]any{
					"size": "50Gi",
				},
				"database": map[string]any{
					"size": "10Gi",
				},
			},
		},
		"harborAdminPassword": "",            // Not used when existingSecret is set
		"existingSecret":      "harbor-admin", // Secret synced by SecretProviderClass + secret-sync pod
		"core": map[string]any{
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "1000m", "memory": "1Gi"},
				"requests": map[string]any{"cpu": "100m", "memory": "256Mi"},
			},
		},
		// NOTE: harbor helm chart hardcodes RollingUpdate strategy. jobservice and registry
		// both use RWO PVCs — if ArgoCD syncs a pod-spec change, the rolling update will
		// deadlock (new pod can't attach PVC held by old pod on a different node).
		// Workaround: kubectl patch deployment harbor-{jobservice,registry} -n harbor \
		//   --type=merge -p '{"spec":{"strategy":{"type":"Recreate","rollingUpdate":null}}}'
		"jobservice": map[string]any{
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
		},
		"registry": map[string]any{
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "1000m", "memory": "1Gi"},
				"requests": map[string]any{"cpu": "100m", "memory": "256Mi"},
			},
		},
	}

	cdk8s.NewHelm(chart, jsii.String("harbor-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("harbor"),
		Repo:        jsii.String("https://helm.goharbor.io"),
		Version:     jsii.String("1.18.2"),
		ReleaseName: jsii.String("harbor"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// Gateway API HTTPRoute — routes harbor.madhan.app → harbor-core:80
	cdk8s.NewApiObject(chart, jsii.String("harbor-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("harbor"),
			Namespace: jsii.String(namespace),
		},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "homelab-gateway", "namespace": "kube-system"},
		},
		"hostnames": []string{"harbor.madhan.app"},
		"rules": []map[string]any{
			{
				"matches": []map[string]any{
					{"path": map[string]any{"type": "PathPrefix", "value": "/"}},
				},
				"backendRefs": []map[string]any{
					{"group": "", "kind": "Service", "name": "harbor-core", "port": 80, "weight": 1},
				},
			},
		},
	}))

	return chart
}
