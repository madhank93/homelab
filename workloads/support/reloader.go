// Package support holds cluster-wide helper controllers.
package support

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
)

// NewReloaderChart deploys Stakater Reloader, which restarts workloads annotated
// reloader.stakater.com/auto: "true" when a ConfigMap or Secret they reference changes.
func NewReloaderChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	reloader := config.HelmChart("reloader")
	cdk8s.NewHelm(chart, jsii.String("reloader"), &cdk8s.HelmProps{
		Chart:       jsii.String("reloader"),
		Repo:        jsii.String(reloader.Repo),
		Version:     jsii.String(reloader.Version),
		ReleaseName: jsii.String("reloader"),
		Namespace:   jsii.String(namespace),
	})

	return chart
}
