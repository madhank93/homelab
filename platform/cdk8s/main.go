package main

import (
	"fmt"

	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/cots/ai"
	"github.com/madhank93/homelab/cdk8s/cots/automation"
	"github.com/madhank93/homelab/cdk8s/cots/management"
	"github.com/madhank93/homelab/cdk8s/cots/monitoring"
	"github.com/madhank93/homelab/cdk8s/cots/registry"
	"github.com/madhank93/homelab/cdk8s/cots/seccomp"
	"github.com/madhank93/homelab/cdk8s/cots/storage"
)

func main() {
	rootFolder := "../../app"

	// 1. Sealed Secrets Controller
	sealedSecretsApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/sealed-secrets", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	seccomp.NewSealedSecretsChart(sealedSecretsApp, "sealed-secrets", "kube-system")
	sealedSecretsApp.Synth()

	// 2. Cert-Manager
	certMgrApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/cert-manager", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	seccomp.NewCertManagerChart(certMgrApp, "cert-manager-app", "cert-manager")
	certMgrApp.Synth()

	// 3. Longhorn Storage
	longhornApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/longhorn", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	storage.NewLonghornChart(longhornApp, "longhorn-app", "longhorn-system")
	longhornApp.Synth()

	// 4. Infisical (includes namespace, backend, frontend, operator, service token)
	infisicalApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/infisical", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	seccomp.NewInfisicalChart(infisicalApp, "infisical-app", "infisical")
	infisicalApp.Synth()
	// Monitoring Stack
	grafanaApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/grafana", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	monitoring.NewGrafanaChart(grafanaApp, "grafana-app", "grafana")
	grafanaApp.Synth()

	victoriaMetricsApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/victoria-metrics", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	monitoring.NewVictoriaMetricsChart(victoriaMetricsApp, "victoria-metrics-app", "victoria-metrics")
	victoriaMetricsApp.Synth()

	victoriaLogsApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/victoria-logs", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	monitoring.NewVictoriaLogsChart(victoriaLogsApp, "victoria-logs-app", "victoria-logs")
	victoriaLogsApp.Synth()

	alertManagerApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/alertmanager", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	monitoring.NewAlertManagerChart(alertManagerApp, "alertmanager-app", "alertmanager")
	alertManagerApp.Synth()

	// Container Registry
	harborApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/harbor", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	registry.NewHarborChart(harborApp, "harbor-app", "harbor")
	harborApp.Synth()

	// Automation
	n8nApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/n8n", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	automation.NewN8nChart(n8nApp, "n8n-app", "n8n")
	n8nApp.Synth()

	// AI/ML
	nvidiaGpuOperatorApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/nvidia-gpu-operator", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	ai.NewNvidiaGpuOperatorChart(nvidiaGpuOperatorApp, "nvidia-gpu-operator", "nvidia-gpu-operator")
	nvidiaGpuOperatorApp.Synth()

	ollamaApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/ollama", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	ai.NewOllamaChart(ollamaApp, "ollama-app", "ollama")
	ollamaApp.Synth()

	// Security & Compliance
	trivyApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/trivy", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	seccomp.NewTrivyChart(trivyApp, "trivy-app", "trivy")
	trivyApp.Synth()

	// Management Tools
	headlampApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/headlamp", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	management.NewHeadlampChart(headlampApp, "headlamp-app", "headlamp")
	headlampApp.Synth()

	fleetApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/fleet", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	management.NewFleetChart(fleetApp, "fleet-app", "fleet")
	fleetApp.Synth()
}
