package gaming

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/kit"
)

const (
	cs2StorageClass = "longhorn-cs2"
	// Fixed so players' `connect` command never changes. Kept out of the
	// .221-.224 node addresses that share the LB pool range.
	cs2LoadBalancerIP = "192.168.1.230"
	cs2Port           = 27015
)

// NewCS2Chart deploys a LAN-only Counter-Strike 2 dedicated server
// (sv_lan 1, so no Steam game server token and no public server browser
// listing). Players join with `connect 192.168.1.230` from the console.
//
// The image is small; steamcmd downloads the ~60 GB game into the PVC on every
// start and updates it in place, so a restart picks up a CS2 update.
func NewCS2Chart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, false)

	kit.SecretProviderClass(chart, "cs2-secrets", "cs2", []string{"RCON_PASSWORD", "JOIN_PASSWORD"}, "cs2-secrets",
		kit.Sync{Object: "RCON_PASSWORD", Key: "RCON_PASSWORD"},
		kit.Sync{Object: "JOIN_PASSWORD", Key: "JOIN_PASSWORD"})

	k8s.NewKubeServiceAccount(chart, jsii.String("cs2-sa"), &k8s.KubeServiceAccountProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("cs2")},
	})

	// One replica: the game files are re-downloadable from Steam, so losing
	// the replica costs one download rather than another ~60 GB of disk.
	k8s.NewKubeStorageClass(chart, jsii.String("cs2-storageclass"), &k8s.KubeStorageClassProps{
		Metadata:             &k8s.ObjectMeta{Name: jsii.String(cs2StorageClass)},
		Provisioner:          jsii.String("driver.longhorn.io"),
		ReclaimPolicy:        jsii.String("Delete"),
		AllowVolumeExpansion: jsii.Bool(true),
		VolumeBindingMode:    jsii.String("Immediate"),
		Parameters: &map[string]*string{
			"numberOfReplicas":    jsii.String("1"),
			"staleReplicaTimeout": jsii.String("30"),
			"fsType":              jsii.String("ext4"),
		},
	})

	storageClass := cs2StorageClass
	k8s.NewKubePersistentVolumeClaim(chart, jsii.String("cs2-data"), &k8s.KubePersistentVolumeClaimProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("cs2-data")},
		Spec: &k8s.PersistentVolumeClaimSpec{
			AccessModes:      &[]*string{jsii.String("ReadWriteOnce")},
			StorageClassName: &storageClass,
			Resources: &k8s.VolumeResourceRequirements{
				// The game is ~60 GB; the headroom absorbs updates, which steamcmd
				// stages next to the installed files.
				Requests: &map[string]k8s.Quantity{"storage": k8s.Quantity_FromString(jsii.String("80Gi"))},
			},
		},
	})

	secretEnv := func(name, key string) *k8s.EnvVar {
		return &k8s.EnvVar{
			Name: jsii.String(name),
			ValueFrom: &k8s.EnvVarSource{
				SecretKeyRef: &k8s.SecretKeySelector{Name: jsii.String("cs2-secrets"), Key: jsii.String(key)},
			},
		}
	}
	env := func(name, value string) *k8s.EnvVar {
		return &k8s.EnvVar{Name: jsii.String(name), Value: jsii.String(value)}
	}

	labels := &map[string]*string{"app": jsii.String("cs2")}
	k8s.NewKubeDeployment(chart, jsii.String("cs2"), &k8s.KubeDeploymentProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("cs2")},
		Spec: &k8s.DeploymentSpec{
			Replicas: jsii.Number(1),
			// Recreate: the RWO game volume can only be attached to one pod.
			Strategy: &k8s.DeploymentStrategy{Type: jsii.String("Recreate")},
			Selector: &k8s.LabelSelector{MatchLabels: labels},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{
					Labels:      labels,
					Annotations: &map[string]*string{"reloader.stakater.com/auto": jsii.String("true")},
				},
				Spec: &k8s.PodSpec{
					ServiceAccountName: jsii.String("cs2"),
					// The image runs as steam (1000:1000) and writes the game into the PVC.
					SecurityContext: &k8s.PodSecurityContext{FsGroup: jsii.Number(1000)},
					Containers: &[]*k8s.Container{{
						Name:  jsii.String("cs2"),
						Image: jsii.String(config.ContainerImage("cs2").Ref()),
						Ports: &[]*k8s.ContainerPort{
							{ContainerPort: jsii.Number(cs2Port), Name: jsii.String("game"), Protocol: jsii.String("UDP")},
							{ContainerPort: jsii.Number(cs2Port), Name: jsii.String("rcon"), Protocol: jsii.String("TCP")},
						},
						Env: &[]*k8s.EnvVar{
							env("CS2_SERVERNAME", "homelab LAN"),
							env("CS2_LAN", "1"),
							env("CS2_GAMEALIAS", "competitive"),
							env("CS2_MAPGROUP", "mg_active"),
							env("CS2_STARTMAP", "de_inferno"),
							env("CS2_MAXPLAYERS", "10"),
							env("CS2_BOT_QUOTA", "0"),
							// Hibernation when empty is a known crash source in this image.
							env("CS2_SERVER_HIBERNATE", "0"),
							secretEnv("CS2_RCONPW", "RCON_PASSWORD"),
							secretEnv("CS2_PW", "JOIN_PASSWORD"),
						},
						Resources: &k8s.ResourceRequirements{
							Requests: &map[string]k8s.Quantity{
								"cpu":    k8s.Quantity_FromString(jsii.String("2000m")),
								"memory": k8s.Quantity_FromString(jsii.String("2Gi")),
							},
							Limits: &map[string]k8s.Quantity{
								"cpu":    k8s.Quantity_FromString(jsii.String("4000m")),
								"memory": k8s.Quantity_FromString(jsii.String("4Gi")),
							},
						},
						VolumeMounts: &[]*k8s.VolumeMount{
							{Name: jsii.String("data"), MountPath: jsii.String("/home/steam/cs2-dedicated")},
							kit.SecretsMount(),
						},
					}},
					Volumes: &[]*k8s.Volume{
						{
							Name:                  jsii.String("data"),
							PersistentVolumeClaim: &k8s.PersistentVolumeClaimVolumeSource{ClaimName: jsii.String("cs2-data")},
						},
						kit.SecretsVolume("cs2-secrets"),
					},
				},
			},
		},
	})

	k8s.NewKubeService(chart, jsii.String("cs2-service"), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name:        jsii.String("cs2"),
			Annotations: &map[string]*string{"lbipam.cilium.io/ips": jsii.String(cs2LoadBalancerIP)},
		},
		Spec: &k8s.ServiceSpec{
			Type:     jsii.String("LoadBalancer"),
			Selector: labels,
			Ports: &[]*k8s.ServicePort{
				{Name: jsii.String("game"), Port: jsii.Number(cs2Port), TargetPort: k8s.IntOrString_FromString(jsii.String("game")), Protocol: jsii.String("UDP")},
				{Name: jsii.String("rcon"), Port: jsii.Number(cs2Port), TargetPort: k8s.IntOrString_FromString(jsii.String("rcon")), Protocol: jsii.String("TCP")},
			},
		},
	})

	return chart
}
