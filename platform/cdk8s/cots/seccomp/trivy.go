package seccomp

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/trivyoperator"
)

func NewTrivyChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	values := map[string]any{
		"trivy-operator": map[string]any{
			"enabled": true,
			"serviceMonitor": map[string]any{
				"enabled": true,
			},
		},
		"operator": map[string]any{
			"replicas":                1,
			"scanJobsConcurrentLimit": 3,
			"scanJobsRetryDelay":      "30s",
		},
		"trivyOperator": map[string]any{
			"scanJobTimeout": "5m",
		},
		"compliance": map[string]any{
			"failedChecksOnly": false,
		},
		"rbac": map[string]any{
			"create": true,
		},
		"serviceAccount": map[string]any{
			"create": true,
			"name":   "trivy-operator",
		},
	}

	trivyoperator.NewTrivyoperator(chart, jsii.String("trivy-release"), &trivyoperator.TrivyoperatorProps{
		ReleaseName: jsii.String("trivy-operator"),
		Values:      &values,
	})

	return chart
}
