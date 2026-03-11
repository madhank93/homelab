package databases

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

func NewCnpgOperatorChart(scope constructs.Construct, id string) cdk8s.Chart {
	namespace := "cnpg-system"
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	k8s.NewKubeNamespace(chart, jsii.String("cnpg-namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(namespace),
		},
	})

	cdk8s.NewHelm(chart, jsii.String("cnpg-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("cloudnative-pg"),
		Repo:        jsii.String("https://cloudnative-pg.github.io/charts"),
		Version:     jsii.String("0.27.1"),
		ReleaseName: jsii.String("cnpg"),
		Namespace:   jsii.String(namespace),
		Values: &map[string]any{
			"monitoring": map[string]any{
				// Creates PodMonitors for each CNPG cluster (discovered by VMAgent's podMonitorSelector)
				"podMonitorEnabled": true,
				// Creates a Grafana dashboard ConfigMap with label grafana_dashboard=1
				"grafanaDashboard": map[string]any{
					"create": true,
				},
			},
		},
	})

	return chart
}
