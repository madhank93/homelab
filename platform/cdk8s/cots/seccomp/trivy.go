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

	values := map[string]interface{}{
		"trivy-operator": map[string]interface{}{
			"enabled": true,
			"serviceMonitor": map[string]interface{}{
				"enabled": true,
			},
		},
		"operator": map[string]interface{}{
			"replicas":                1,
			"scanJobsConcurrentLimit": 3,
			"scanJobsRetryDelay":      "30s",
		},
		"trivyOperator": map[string]interface{}{
			"scanJobTimeout": "5m",
		},
		"compliance": map[string]interface{}{
			"failedChecksOnly": false,
		},
		"rbac": map[string]interface{}{
			"create": true,
		},
		"serviceAccount": map[string]interface{}{
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
