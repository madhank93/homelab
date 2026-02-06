package storage

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/longhorn"
)

func NewLonghornChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String("longhorn-system"),
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

	longhorn.NewLonghorn(chart, jsii.String("longhorn"), &longhorn.LonghornProps{
		ReleaseName: jsii.String("longhorn"),
		Values:      &values,
	})

	return chart
}
