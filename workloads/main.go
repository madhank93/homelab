package main

import (
	"fmt"

	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/ai"
	"github.com/madhank93/homelab/workloads/automation"
	"github.com/madhank93/homelab/workloads/hardware"
	"github.com/madhank93/homelab/workloads/management"
	"github.com/madhank93/homelab/workloads/monitoring"
	"github.com/madhank93/homelab/workloads/networking"
	"github.com/madhank93/homelab/workloads/observability"
	"github.com/madhank93/homelab/workloads/registry"
	"github.com/madhank93/homelab/workloads/secrets"
	"github.com/madhank93/homelab/workloads/security"
	"github.com/madhank93/homelab/workloads/storage"
	"github.com/madhank93/homelab/workloads/support"
)

func main() {
	rootFolder := "../app"

	// 1. Longhorn Storage
	longhornApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/longhorn", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	storage.NewLonghornChart(longhornApp, "longhorn-app", "longhorn-system")
	longhornApp.Synth()

	// Infisical (includes namespace, backend, frontend, operator, service token)
	infisicalApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/infisical", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	secrets.NewInfisicalChart(infisicalApp, "infisical-app", "infisical")
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
	observability.NewVictoriaMetricsChart(victoriaMetricsApp, "victoria-metrics-app", "victoria-metrics")
	victoriaMetricsApp.Synth()

	victoriaLogsApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/victoria-logs", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	observability.NewVictoriaLogsChart(victoriaLogsApp, "victoria-logs-app", "victoria-logs")
	victoriaLogsApp.Synth()

	alertManagerApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/alertmanager", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	observability.NewAlertManagerChart(alertManagerApp, "alertmanager-app", "alertmanager")
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
	hardware.NewNvidiaGpuOperatorChart(nvidiaGpuOperatorApp, "nvidia-gpu-operator", "nvidia-gpu-operator")
	nvidiaGpuOperatorApp.Synth()

	ollamaApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/ollama", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	ai.NewOllamaChart(ollamaApp, "ollama-app", "ollama")
	ollamaApp.Synth()

	comfyuiApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/comfyui", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	ai.NewComfyUIChart(comfyuiApp, "comfyui-app", "comfyui")
	comfyuiApp.Synth()

	// Security & Compliance
	trivyApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/trivy", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	security.NewTrivyChart(trivyApp, "trivy-app", "trivy")
	trivyApp.Synth()

	falcoApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/falco", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	security.NewFalcoChart(falcoApp, "falco-app", "falco")
	falcoApp.Synth()

	// Observability
	otelApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/opentelemetry", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	observability.NewOtelCollectorChart(otelApp, "otel-app", "opentelemetry")
	otelApp.Synth()

	// NetBird routing peer — advertises 192.168.1.0/24 into WireGuard mesh
	netbirdApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/netbird", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	networking.NewNetbirdPeerChart(netbirdApp, "netbird-app", "netbird")
	netbirdApp.Synth()

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

	rancherApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/rancher", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	management.NewRancherChart(rancherApp, "rancher-app", "cattle-system")
	rancherApp.Synth()

	// Stakater Reloader
	reloaderApp := cdk8s.NewApp(&cdk8s.AppProps{
		Outdir:         jsii.String(fmt.Sprintf("%s/reloader", rootFolder)),
		YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
	})
	support.NewReloaderChart(reloaderApp, "reloader-app", "reloader")
	reloaderApp.Synth()
}
