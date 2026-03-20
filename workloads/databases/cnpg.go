package databases

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

// NewCnpgOperatorChart deploys the CloudNativePG operator (chart cloudnative-pg v0.27.1)
// into the cnpg-system namespace.
//
// CloudNativePG manages PostgreSQL cluster lifecycles including automated failover,
// backup, and restore. Individual database clusters (e.g. for n8n) are created as
// CNPG Cluster CRs in their respective application namespaces.
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
	})

	return chart
}
