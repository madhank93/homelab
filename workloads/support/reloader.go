package support

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

type ReloaderChart struct {
	cdk8s.Chart
}

func NewReloaderChart(scope constructs.Construct, id string, namespace string) *ReloaderChart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	cdk8s.NewHelm(chart, jsii.String("reloader"), &cdk8s.HelmProps{
		Chart:       jsii.String("reloader"),
		Repo:        jsii.String("https://stakater.github.io/stakater-charts"),
		Version:     jsii.String("2.2.8"),
		ReleaseName: jsii.String("reloader"),
		Namespace:   jsii.String(namespace),
	})

	return &ReloaderChart{chart}
}
