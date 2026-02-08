package storage

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewLonghornChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{
		"defaultSettings": map[string]any{
			"defaultReplicaCount": 3,
			"defaultDataPath":     "/var/lib/longhorn/",
		},
		"persistence": map[string]any{
			"defaultClass": true,
		},
	}

	cdk8s.NewHelm(chart, jsii.String("longhorn-release"), &cdk8s.HelmProps{
		Chart:       jsii.String("longhorn"),
		Repo:        jsii.String("https://charts.longhorn.io"),
		Version:     jsii.String("1.11.0"),
		ReleaseName: jsii.String("longhorn"),
		Namespace:   jsii.String(namespace),
		Values:      &values,
	})

	return chart
}
