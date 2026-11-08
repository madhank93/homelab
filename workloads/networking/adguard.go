package networking

import (
	_ "embed"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/kit"
)

const (
	adguardUIPort      = 3000
	adguardMetricsPort = 9618
	adguardConfDir     = "/opt/adguardhome/conf"
)

//go:embed adguard_config.yml
var adguardConfig string

// adguardDashboard is the exporter's upstream dashboard with its datasource
// input replaced by the default (VictoriaMetrics).
//
//go:embed adguard_dashboard.json
var adguardDashboard string

// NewAdGuardChart deploys AdGuard Home as LAN DNS with ad and tracker blocking:
// two pods on different nodes behind one fixed LoadBalancer IP, so a node
// reboot does not take DNS down. Config lives in adguard_config.yml.
//
// Login sessions and query logs are per pod, so the UI at adguard.<domain> is
// served by adguard-0 only; DNS uses both pods.
func NewAdGuardChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, false)

	k8s.NewKubeConfigMap(chart, jsii.String("adguard-config"), &k8s.KubeConfigMapProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("adguard-config")},
		Data:     &map[string]*string{"AdGuardHome.yaml": jsii.String(adguardConfig)},
	})

	// File-only: the init container reads the hash, the exporter the password.
	kit.SecretProviderClass(chart, "adguard-secrets", "adguard", []string{"ADMIN_PASSWORD", "ADMIN_PASSWORD_HASH"}, "")

	k8s.NewKubeServiceAccount(chart, jsii.String("adguard-sa"), &k8s.KubeServiceAccountProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("adguard")},
	})

	image := jsii.String(config.ContainerImage("adguardhome").Ref())
	confMount := &k8s.VolumeMount{Name: jsii.String("conf"), MountPath: jsii.String(adguardConfDir)}
	labels := &map[string]*string{"app": jsii.String("adguard")}
	k8s.NewKubeStatefulSet(chart, jsii.String("adguard"), &k8s.KubeStatefulSetProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("adguard")},
		Spec: &k8s.StatefulSetSpec{
			Replicas:            jsii.Number(2),
			ServiceName:         jsii.String("adguard"),
			PodManagementPolicy: jsii.String("Parallel"),
			Selector:            &k8s.LabelSelector{MatchLabels: labels},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{
					Labels:      labels,
					Annotations: &map[string]*string{"reloader.stakater.com/auto": jsii.String("true")},
				},
				Spec: &k8s.PodSpec{
					ServiceAccountName: jsii.String("adguard"),
					// The image grants its binary cap_net_bind_service, so port 53 works as nobody.
					SecurityContext: &k8s.PodSecurityContext{
						RunAsNonRoot: jsii.Bool(true),
						RunAsUser:    jsii.Number(65534),
						RunAsGroup:   jsii.Number(65534),
						FsGroup:      jsii.Number(65534),
					},
					Affinity: &k8s.Affinity{PodAntiAffinity: &k8s.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &[]*k8s.PodAffinityTerm{{
							LabelSelector: &k8s.LabelSelector{MatchLabels: labels},
							TopologyKey:   jsii.String("kubernetes.io/hostname"),
						}},
					}},
					InitContainers: &[]*k8s.Container{{
						Name:    jsii.String("render-config"),
						Image:   image,
						Command: &[]*string{jsii.String("sh"), jsii.String("-c")},
						Args: &[]*string{jsii.String(
							`sed "s|__ADMIN_PASSWORD_HASH__|$(cat /mnt/secrets/ADMIN_PASSWORD_HASH)|" /template/AdGuardHome.yaml > ` +
								adguardConfDir + `/AdGuardHome.yaml`)},
						VolumeMounts: &[]*k8s.VolumeMount{
							confMount,
							{Name: jsii.String("template"), MountPath: jsii.String("/template")},
							kit.SecretsMount(),
						},
					}},
					Containers: &[]*k8s.Container{
						{
							Name:  jsii.String("adguard"),
							Image: image,
							Ports: &[]*k8s.ContainerPort{
								{ContainerPort: jsii.Number(53), Name: jsii.String("dns-udp"), Protocol: jsii.String("UDP")},
								{ContainerPort: jsii.Number(53), Name: jsii.String("dns-tcp"), Protocol: jsii.String("TCP")},
								{ContainerPort: jsii.Number(adguardUIPort), Name: jsii.String("http")},
							},
							ReadinessProbe: &k8s.Probe{
								TcpSocket:     &k8s.TcpSocketAction{Port: k8s.IntOrString_FromString(jsii.String("dns-tcp"))},
								PeriodSeconds: jsii.Number(10),
							},
							LivenessProbe: &k8s.Probe{
								TcpSocket:           &k8s.TcpSocketAction{Port: k8s.IntOrString_FromString(jsii.String("dns-tcp"))},
								InitialDelaySeconds: jsii.Number(30),
								PeriodSeconds:       jsii.Number(20),
							},
							Resources: &k8s.ResourceRequirements{
								Requests: &map[string]k8s.Quantity{
									"cpu":    k8s.Quantity_FromString(jsii.String("50m")),
									"memory": k8s.Quantity_FromString(jsii.String("128Mi")),
								},
								Limits: &map[string]k8s.Quantity{
									"memory": k8s.Quantity_FromString(jsii.String("512Mi")),
								},
							},
							VolumeMounts: &[]*k8s.VolumeMount{
								confMount,
								{Name: jsii.String("work"), MountPath: jsii.String("/opt/adguardhome/work")},
							},
						},
						{
							Name:  jsii.String("exporter"),
							Image: jsii.String(config.ContainerImage("adguard-exporter").Ref()),
							Env: &[]*k8s.EnvVar{
								{Name: jsii.String("ADGUARD_SERVERS"), Value: jsii.String("http://localhost:3000")},
								{Name: jsii.String("ADGUARD_USERNAMES"), Value: jsii.String("admin")},
								{Name: jsii.String("ADGUARD_PASSWORDS_FILE"), Value: jsii.String("/mnt/secrets/ADMIN_PASSWORD")},
							},
							Ports: &[]*k8s.ContainerPort{
								{ContainerPort: jsii.Number(adguardMetricsPort), Name: jsii.String("metrics")},
							},
							Resources: &k8s.ResourceRequirements{
								Requests: &map[string]k8s.Quantity{
									"cpu":    k8s.Quantity_FromString(jsii.String("10m")),
									"memory": k8s.Quantity_FromString(jsii.String("32Mi")),
								},
								Limits: &map[string]k8s.Quantity{
									"memory": k8s.Quantity_FromString(jsii.String("64Mi")),
								},
							},
							VolumeMounts: &[]*k8s.VolumeMount{kit.SecretsMount()},
						},
					},
					Volumes: &[]*k8s.Volume{
						{Name: jsii.String("conf"), EmptyDir: &k8s.EmptyDirVolumeSource{}},
						{Name: jsii.String("work"), EmptyDir: &k8s.EmptyDirVolumeSource{}},
						{Name: jsii.String("template"), ConfigMap: &k8s.ConfigMapVolumeSource{Name: jsii.String("adguard-config")}},
						kit.SecretsVolume("adguard-secrets"),
					},
				},
			},
		},
	})

	// Governing headless Service for the StatefulSet; also the metrics target.
	k8s.NewKubeService(chart, jsii.String("adguard-headless"), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("adguard"), Labels: labels},
		Spec: &k8s.ServiceSpec{
			ClusterIp: jsii.String("None"),
			Selector:  labels,
			Ports: &[]*k8s.ServicePort{
				{Name: jsii.String("metrics"), Port: jsii.Number(adguardMetricsPort), TargetPort: k8s.IntOrString_FromString(jsii.String("metrics"))},
			},
		},
	})

	k8s.NewKubeService(chart, jsii.String("adguard-dns"), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name:        jsii.String("adguard-dns"),
			Annotations: &map[string]*string{"lbipam.cilium.io/ips": jsii.String(config.C.Cluster.LBIPs.AdGuard)},
		},
		Spec: &k8s.ServiceSpec{
			Type:     jsii.String("LoadBalancer"),
			Selector: labels,
			Ports: &[]*k8s.ServicePort{
				{Name: jsii.String("dns-udp"), Port: jsii.Number(53), TargetPort: k8s.IntOrString_FromString(jsii.String("dns-udp")), Protocol: jsii.String("UDP")},
				{Name: jsii.String("dns-tcp"), Port: jsii.Number(53), TargetPort: k8s.IntOrString_FromString(jsii.String("dns-tcp")), Protocol: jsii.String("TCP")},
			},
		},
	})

	k8s.NewKubeService(chart, jsii.String("adguard-ui"), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("adguard-ui")},
		Spec: &k8s.ServiceSpec{
			Selector: &map[string]*string{"statefulset.kubernetes.io/pod-name": jsii.String("adguard-0")},
			Ports: &[]*k8s.ServicePort{
				{Name: jsii.String("http"), Port: jsii.Number(adguardUIPort), TargetPort: k8s.IntOrString_FromString(jsii.String("http"))},
			},
		},
	})
	kit.Route(chart, "adguard", "adguard", "adguard-ui", adguardUIPort)

	cdk8s.NewApiObject(chart, jsii.String("adguard-sm"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("monitoring.coreos.com/v1"),
		Kind:       jsii.String("ServiceMonitor"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String("adguard")},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"selector":  map[string]any{"matchLabels": map[string]any{"app": "adguard"}},
		"endpoints": []map[string]any{{"port": "metrics", "interval": "30s"}},
	}))

	k8s.NewKubeConfigMap(chart, jsii.String("adguard-dashboard"), &k8s.KubeConfigMapProps{
		Metadata: &k8s.ObjectMeta{
			Name:   jsii.String("adguard-dashboard"),
			Labels: &map[string]*string{"grafana_dashboard": jsii.String("1")},
		},
		Data: &map[string]*string{"adguard.json": jsii.String(adguardDashboard)},
	})

	return chart
}
