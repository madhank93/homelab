package support

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

// NewReloaderChart deploys Stakater Reloader, which watches ConfigMaps and Secrets
// for changes and rolls the Deployments/StatefulSets that reference them via the
// reloader.stakater.com/auto: "true" annotation.
func NewReloaderChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
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

	return chart
}
