package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewOpenWebUIChart deploys Open WebUI, a chat front end for the in-cluster Ollama.
//
// It is LAN only: the route serves chat.<domain>, which has no public DNS record.
// The first account to sign up becomes admin; later sign-ups wait as pending.
func NewOpenWebUIChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, false)

	// PVC for the SQLite database, uploads and the generated session secret key
	k8s.NewKubePersistentVolumeClaim(chart, jsii.String("open-webui-data"), &k8s.KubePersistentVolumeClaimProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("open-webui-data"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.PersistentVolumeClaimSpec{
			AccessModes:      &[]*string{jsii.String("ReadWriteOnce")},
			StorageClassName: jsii.String("longhorn"),
			Resources: &k8s.VolumeResourceRequirements{
				Requests: &map[string]k8s.Quantity{
					"storage": k8s.Quantity_FromString(jsii.String("10Gi")),
				},
			},
		},
	})

	replicas := float64(1)
	k8s.NewKubeDeployment(chart, jsii.String("open-webui"), &k8s.KubeDeploymentProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("open-webui"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.DeploymentSpec{
			Replicas: &replicas,
			// Recreate: the RWO volume cannot attach to two pods at once
			Strategy: &k8s.DeploymentStrategy{Type: jsii.String("Recreate")},
			Selector: &k8s.LabelSelector{
				MatchLabels: &map[string]*string{"app": jsii.String("open-webui")},
			},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{
					Labels: &map[string]*string{"app": jsii.String("open-webui")},
				},
				Spec: &k8s.PodSpec{
					Containers: &[]*k8s.Container{
						{
							Name:            jsii.String("open-webui"),
							Image:           jsii.String(config.ContainerImage("open-webui").Ref()),
							ImagePullPolicy: jsii.String("IfNotPresent"),
							Ports: &[]*k8s.ContainerPort{
								{ContainerPort: jsii.Number(8080), Name: jsii.String("http"), Protocol: jsii.String("TCP")},
							},
							Env: &[]*k8s.EnvVar{
								{Name: jsii.String("OLLAMA_BASE_URL"), Value: jsii.String("http://ollama.ollama.svc.cluster.local:11434")},
								// New sign-ups wait for the admin to approve them; only the first account is admin.
								{Name: jsii.String("DEFAULT_USER_ROLE"), Value: jsii.String("pending")},
							},
							Resources: &k8s.ResourceRequirements{
								Limits: &map[string]k8s.Quantity{
									"cpu":    k8s.Quantity_FromString(jsii.String("2000m")),
									"memory": k8s.Quantity_FromString(jsii.String("2Gi")),
								},
								Requests: &map[string]k8s.Quantity{
									"cpu":    k8s.Quantity_FromString(jsii.String("250m")),
									"memory": k8s.Quantity_FromString(jsii.String("512Mi")),
								},
							},
							VolumeMounts: &[]*k8s.VolumeMount{
								{Name: jsii.String("data"), MountPath: jsii.String("/app/backend/data")},
							},
						},
					},
					Volumes: &[]*k8s.Volume{
						{
							Name: jsii.String("data"),
							PersistentVolumeClaim: &k8s.PersistentVolumeClaimVolumeSource{
								ClaimName: jsii.String("open-webui-data"),
							},
						},
					},
				},
			},
		},
	})

	k8s.NewKubeService(chart, jsii.String("open-webui-service"), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("open-webui"),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.ServiceSpec{
			Selector: &map[string]*string{"app": jsii.String("open-webui")},
			Ports: &[]*k8s.ServicePort{
				{
					Name:       jsii.String("http"),
					Port:       jsii.Number(8080),
					TargetPort: k8s.IntOrString_FromNumber(jsii.Number(8080)),
					Protocol:   jsii.String("TCP"),
				},
			},
		},
	})

	kit.Route(chart, "open-webui", "chat", "open-webui", 8080)

	return chart
}
