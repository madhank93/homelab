package registry

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	harborimport "github.com/madhank93/homelab/workloads/imports/harbor"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

func NewHarborChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("harbor-namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
		},
	})

	// SecretProviderClass — Pattern B (secretObjects sync).
	// Fetches HARBOR_ADMIN_PASSWORD from OpenBao and syncs it into k8s Secret "harbor-admin".
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
		"secretObjects": []map[string]any{
			{
				"secretName": "harbor-admin",
				"type":       "Opaque",
				"data": []map[string]any{
					{"objectName": "HARBOR_ADMIN_PASSWORD", "key": "HARBOR_ADMIN_PASSWORD"},
				},
			},
		},
	}))

	k8s.NewKubeServiceAccount(chart, jsii.String("harbor-secret-sync-sa"), &k8s.KubeServiceAccountProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("secret-sync"),
			Namespace: jsii.String(namespace),
		},
	})

	// Secret-sync Deployment — mounts the CSI volume to trigger secretObjects sync.
	// Harbor's Helm chart does not support extraVolumes on its component pods.
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
					Labels:      &map[string]*string{"app": jsii.String("secret-sync")},
					Annotations: &map[string]*string{"reloader.stakater.com/auto": jsii.String("true")},
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

	values := map[string]interface{}{
		"expose": map[string]interface{}{
			"type": "clusterIP",
			"tls":  map[string]interface{}{"enabled": false},
		},
		"externalURL": "https://harbor.madhan.app",
		"persistence": map[string]interface{}{
			"enabled": true,
			// ReadWriteMany via Longhorn NFS — eliminates RWO multi-attach deadlocks
			// on rolling updates. Longhorn automatically provisions an NFS server pod
			// for each RWX volume.
			"persistentVolumeClaim": map[string]interface{}{
				"registry": map[string]interface{}{
					"size":       "50Gi",
					"accessMode": "ReadWriteMany",
				},
				"jobservice": map[string]interface{}{
					"jobLog": map[string]interface{}{
						"accessMode": "ReadWriteMany",
					},
				},
				"database": map[string]interface{}{"size": "10Gi"},
			},
		},
		"harborAdminPassword": "",
		"existingSecret":      "harbor-admin", // Secret synced by SecretProviderClass + secret-sync pod
		"core": map[string]interface{}{
			"resources": map[string]interface{}{
				"limits":   map[string]interface{}{"cpu": "1000m", "memory": "1Gi"},
				"requests": map[string]interface{}{"cpu": "100m", "memory": "256Mi"},
			},
		},
		"jobservice": map[string]interface{}{
			"resources": map[string]interface{}{
				"limits":   map[string]interface{}{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]interface{}{"cpu": "100m", "memory": "128Mi"},
			},
		},
		"registry": map[string]interface{}{
			"resources": map[string]interface{}{
				"limits":   map[string]interface{}{"cpu": "1000m", "memory": "1Gi"},
				"requests": map[string]interface{}{"cpu": "100m", "memory": "256Mi"},
			},
		},
	}

	harborimport.NewHarbor(chart, jsii.String("harbor-release"), &harborimport.HarborProps{
		ReleaseName: jsii.String("harbor"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	// Gateway API HTTPRoute — routes harbor.madhan.app → harbor:80 (nginx proxy)
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
					{"group": "", "kind": "Service", "name": "harbor", "port": 80, "weight": 1},
				},
			},
		},
	}))

	return chart
}
