package databases

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/kit"
)

// NewCnpgOperatorChart deploys the CloudNativePG operator into cnpg-system.
// Database clusters are CNPG Cluster CRs in each application's namespace.
func NewCnpgOperatorChart(scope constructs.Construct, id string) cdk8s.Chart {
	namespace := "cnpg-system"
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	kit.Namespace(chart, namespace, false)

	cnpg := config.HelmChart("cloudnative-pg")
	cdk8s.NewHelm(chart, jsii.String("cnpg-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("cloudnative-pg"),
		Repo:        jsii.String(cnpg.Repo),
		Version:     jsii.String(cnpg.Version),
		ReleaseName: jsii.String("cnpg"),
		Namespace:   jsii.String(namespace),
	})

	return chart
}
