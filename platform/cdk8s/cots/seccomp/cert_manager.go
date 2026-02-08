package seccomp

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/certmanager"
)

func NewCertManagerChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// Install Cert-Manager via Helm
	// Note: ClusterIssuers and Certificates should be created separately
	// See: platform/cdk8s/manifests/cert-manager-issuers.yaml (to be created)
	certmanager.NewCertmanager(chart, jsii.String("cert-manager"), &certmanager.CertmanagerProps{
		ReleaseName: jsii.String("cert-manager"),
		Namespace:   jsii.String(namespace),
		Values: &certmanager.HelmValues{
			InstallCrDs: jsii.Bool(true),
			Global: &certmanager.HelmValuesGlobal{
				LeaderElection: &certmanager.HelmValuesGlobalLeaderElection{
					Namespace: jsii.String(namespace),
				},
			},
		},
	})

	return chart
}
