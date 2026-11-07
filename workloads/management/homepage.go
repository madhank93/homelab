package management

import (
	"embed"
	"path"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/k8s"
	"github.com/madhank93/homelab/workloads/kit"
)

// Homepage's config files; each becomes a key of the homepage ConfigMap.
//
//go:embed homepage/*
var homepageConfig embed.FS

// NewHomepageChart deploys Homepage, the start page linking every homelab app.
// Login goes through Authentik OIDC, limited to the grafana-admins group.
// The config lives in workloads/management/homepage/.
func NewHomepageChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, false)

	files, err := homepageConfig.ReadDir("homepage")
	if err != nil {
		panic(err)
	}
	data := map[string]*string{}
	for _, f := range files {
		body, err := homepageConfig.ReadFile(path.Join("homepage", f.Name()))
		if err != nil {
			panic(err)
		}
		data[f.Name()] = jsii.String(string(body))
	}
	k8s.NewKubeConfigMap(chart, jsii.String("homepage-config"), &k8s.KubeConfigMapProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("homepage")},
		Data:     &data,
	})

	kit.SecretProviderClass(chart, "homepage-secrets", "homepage", []string{"AUTH_SECRET", "OAUTH_CLIENT_SECRET"}, "homepage-secrets",
		kit.Sync{Object: "AUTH_SECRET", Key: "AUTH_SECRET"},
		kit.Sync{Object: "OAUTH_CLIENT_SECRET", Key: "OAUTH_CLIENT_SECRET"})

	// Read-only cluster access for the kubernetes widget (cluster and node usage).
	k8s.NewKubeServiceAccount(chart, jsii.String("homepage-sa"), &k8s.KubeServiceAccountProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("homepage")},
	})
	k8s.NewKubeClusterRole(chart, jsii.String("homepage-clusterrole"), &k8s.KubeClusterRoleProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("homepage")},
		Rules: &[]*k8s.PolicyRule{
			{
				ApiGroups: &[]*string{jsii.String("")},
				Resources: &[]*string{jsii.String("namespaces"), jsii.String("pods"), jsii.String("nodes")},
				Verbs:     &[]*string{jsii.String("get"), jsii.String("list")},
			},
			{
				ApiGroups: &[]*string{jsii.String("metrics.k8s.io")},
				Resources: &[]*string{jsii.String("nodes"), jsii.String("pods")},
				Verbs:     &[]*string{jsii.String("get"), jsii.String("list")},
			},
		},
	})
	k8s.NewKubeClusterRoleBinding(chart, jsii.String("homepage-crb"), &k8s.KubeClusterRoleBindingProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("homepage")},
		RoleRef: &k8s.RoleRef{
			ApiGroup: jsii.String("rbac.authorization.k8s.io"),
			Kind:     jsii.String("ClusterRole"),
			Name:     jsii.String("homepage"),
		},
		Subjects: &[]*k8s.Subject{
			{Kind: jsii.String("ServiceAccount"), Name: jsii.String("homepage"), Namespace: jsii.String(namespace)},
		},
	})

	secretEnv := func(name, key string) *k8s.EnvVar {
		return &k8s.EnvVar{
			Name: jsii.String(name),
			ValueFrom: &k8s.EnvVarSource{
				SecretKeyRef: &k8s.SecretKeySelector{Name: jsii.String("homepage-secrets"), Key: jsii.String(key)},
			},
		}
	}
	env := func(name, value string) *k8s.EnvVar {
		return &k8s.EnvVar{Name: jsii.String(name), Value: jsii.String(value)}
	}

	// One subPath mount per config file keeps /app/config writable for the logs dir.
	mounts := []*k8s.VolumeMount{
		{Name: jsii.String("logs"), MountPath: jsii.String("/app/config/logs")},
		kit.SecretsMount(),
	}
	for _, f := range files {
		mounts = append(mounts, &k8s.VolumeMount{
			Name:      jsii.String("config"),
			MountPath: jsii.String("/app/config/" + f.Name()),
			SubPath:   jsii.String(f.Name()),
		})
	}

	host := config.C.Host("home")
	labels := &map[string]*string{"app": jsii.String("homepage")}
	k8s.NewKubeDeployment(chart, jsii.String("homepage"), &k8s.KubeDeploymentProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String("homepage"),
			// subPath mounts never see ConfigMap updates, so restart on change.
			Annotations: &map[string]*string{"reloader.stakater.com/auto": jsii.String("true")},
		},
		Spec: &k8s.DeploymentSpec{
			Replicas: jsii.Number(1),
			Selector: &k8s.LabelSelector{MatchLabels: labels},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{Labels: labels},
				Spec: &k8s.PodSpec{
					ServiceAccountName: jsii.String("homepage"),
					Containers: &[]*k8s.Container{{
						Name:  jsii.String("homepage"),
						Image: jsii.String(config.ContainerImage("homepage").Ref()),
						Ports: &[]*k8s.ContainerPort{
							{ContainerPort: jsii.Number(3000), Name: jsii.String("http"), Protocol: jsii.String("TCP")},
						},
						Env: &[]*k8s.EnvVar{
							{Name: jsii.String("MY_POD_IP"), ValueFrom: &k8s.EnvVarSource{
								FieldRef: &k8s.ObjectFieldSelector{FieldPath: jsii.String("status.podIP")},
							}},
							// The pod IP entry lets the kubelet probe pass the host check.
							env("HOMEPAGE_ALLOWED_HOSTS", "$(MY_POD_IP):3000,"+host),
							env("HOMEPAGE_VAR_DOMAIN", config.C.Domain),
							env("HOMEPAGE_AUTH_ENABLED", "true"),
							env("HOMEPAGE_EXTERNAL_URL", "https://"+host),
							env("HOMEPAGE_OIDC_ISSUER", "https://"+config.C.Host("auth")+"/application/o/homepage/"),
							env("HOMEPAGE_OIDC_CLIENT_ID", "homepage-homelab"),
							env("HOMEPAGE_OIDC_NAME", "Authentik"),
							env("HOMEPAGE_OIDC_AUTO_LOGIN", "true"),
							secretEnv("HOMEPAGE_AUTH_SECRET", "AUTH_SECRET"),
							secretEnv("HOMEPAGE_OIDC_CLIENT_SECRET", "OAUTH_CLIENT_SECRET"),
						},
						ReadinessProbe: &k8s.Probe{
							HttpGet: &k8s.HttpGetAction{
								Path: jsii.String("/api/healthcheck"),
								Port: k8s.IntOrString_FromString(jsii.String("http")),
							},
							PeriodSeconds: jsii.Number(10),
						},
						SecurityContext: &k8s.SecurityContext{
							AllowPrivilegeEscalation: jsii.Bool(false),
							Capabilities:             &k8s.Capabilities{Drop: &[]*string{jsii.String("ALL")}},
							RunAsNonRoot:             jsii.Bool(true),
							RunAsUser:                jsii.Number(1000),
							RunAsGroup:               jsii.Number(1000),
						},
						Resources: &k8s.ResourceRequirements{
							Limits: &map[string]k8s.Quantity{
								"cpu":    k8s.Quantity_FromString(jsii.String("500m")),
								"memory": k8s.Quantity_FromString(jsii.String("512Mi")),
							},
							Requests: &map[string]k8s.Quantity{
								"cpu":    k8s.Quantity_FromString(jsii.String("50m")),
								"memory": k8s.Quantity_FromString(jsii.String("128Mi")),
							},
						},
						VolumeMounts: &mounts,
					}},
					Volumes: &[]*k8s.Volume{
						{Name: jsii.String("config"), ConfigMap: &k8s.ConfigMapVolumeSource{Name: jsii.String("homepage")}},
						{Name: jsii.String("logs"), EmptyDir: &k8s.EmptyDirVolumeSource{}},
						kit.SecretsVolume("homepage-secrets"),
					},
				},
			},
		},
	})

	k8s.NewKubeService(chart, jsii.String("homepage-service"), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("homepage")},
		Spec: &k8s.ServiceSpec{
			Selector: labels,
			Ports: &[]*k8s.ServicePort{
				{Name: jsii.String("http"), Port: jsii.Number(3000), TargetPort: k8s.IntOrString_FromNumber(jsii.Number(3000)), Protocol: jsii.String("TCP")},
			},
		},
	})

	kit.Route(chart, "homepage", "home", "homepage", 3000)

	return chart
}
