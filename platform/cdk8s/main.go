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
)

func main() {
	rootFolder := "../../app"

	// Cert-Manager
	certMgrApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/cert-manager", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	seccomp.NewCertManagerChart(certMgrApp, "cert-manager-app", "cert-manager")
	certMgrApp.Synth()

	// NVIDIA GPU Operator
	nvidiaGpuOperatorApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/nvidia-gpu-operator", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	ai.NewNvidiaGpuOperatorChart(nvidiaGpuOperatorApp, "nvidia-gpu-operator", "nvidia-gpu-operator")
	nvidiaGpuOperatorApp.Synth()

	// Grafana
	grafanaApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/grafana", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	monitoring.NewGrafanaChart(grafanaApp, "grafana-app", "grafana")
	grafanaApp.Synth()

	// Harbor
	harborApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/harbor", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	registry.NewHarborChart(harborApp, "harbor-app", "harbor")
	harborApp.Synth()

	// Ollama
	ollamaApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/ollama", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	ai.NewOllamaChart(ollamaApp, "ollama-app", "ollama")
	ollamaApp.Synth()

	// Victoria Metrics
	victoriaMetricsApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/victoria-metrics", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	monitoring.NewVictoriaMetricsChart(victoriaMetricsApp, "victoria-metrics-app", "victoria-metrics")
	victoriaMetricsApp.Synth()

	// Victoria Logs
	victoriaLogsApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/victoria-logs", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	monitoring.NewVictoriaLogsChart(victoriaLogsApp, "victoria-logs-app", "victoria-logs")
	victoriaLogsApp.Synth()

	// Alert Manager
	alertManagerApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/alertmanager", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	monitoring.NewAlertManagerChart(alertManagerApp, "alertmanager-app", "alertmanager")
	alertManagerApp.Synth()

	// N8N
	n8nApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/n8n", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	automation.NewN8nChart(n8nApp, "n8n-app", "n8n")
	n8nApp.Synth()

	// Trivy
	trivyApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/trivy", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	seccomp.NewTrivyChart(trivyApp, "trivy-app", "trivy")
	trivyApp.Synth()

	// Infisical
	infisicalApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/infisical", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	seccomp.NewInfisicalChart(infisicalApp, "infisical-app", "infisical")
	infisicalApp.Synth()

	// // Rancher
	// rancherApp := cdk8s.NewApp(&cdk8s.AppProps{
	// 	Outdir:         jsii.String(fmt.Sprintf("%s/rancher", rootFolder)),
	// 	YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	// })
	// management.NewRancherChart(rancherApp, "rancher-app", "rancher")
	// rancherApp.Synth()

	// Headlamp
	headlampApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/headlamp", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	management.NewHeadlampChart(headlampApp, "headlamp-app", "headlamp")
	headlampApp.Synth()

	// Fleet
	fleetApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/fleet", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	management.NewFleetChart(fleetApp, "fleet-app", "fleet")
	fleetApp.Synth()
}
