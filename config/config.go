// Package config exposes the repo's shared infrastructure values (config.yml)
// and version pins (versions.yaml) to both Go programs. The files are embedded,
// so results do not depend on the working directory.
package config

import (
	_ "embed"
	"fmt"

	"go.yaml.in/yaml/v3"
)

//go:embed config.yml
var configYAML []byte

//go:embed versions.yaml
var versionsYAML []byte

// Node is one Talos VM: its LAN address, role and Proxmox sizing.
type Node struct {
	Name     string   `yaml:"name"`
	IP       string   `yaml:"ip"`
	Role     string   `yaml:"role"`
	Cores    int      `yaml:"cores"`
	Memory   int      `yaml:"memory"`
	Disk     int      `yaml:"disk"`
	CPUUnits int      `yaml:"cpu_units"`
	Balloon  int      `yaml:"balloon"`
	PcieIDs  []string `yaml:"pcie_ids"`
}

// HasGPU reports whether the node gets PCIe devices passed through.
func (n Node) HasGPU() bool { return len(n.PcieIDs) > 0 }

// ClusterConfig describes the Talos cluster network and its nodes.
type ClusterConfig struct {
	Name        string                       `yaml:"name"`
	VIP         string                       `yaml:"vip"`
	Subnet      string                       `yaml:"subnet"`
	Router      string                       `yaml:"router"`
	Nameservers []string                     `yaml:"nameservers"`
	GatewayIP   string                       `yaml:"gateway_ip"`
	LBPool      struct{ Start, Stop string } `yaml:"lb_pool"`
	LBIPs       struct {
		CS2    string `yaml:"cs2"`
		Ollama string `yaml:"ollama"`
	} `yaml:"lb_ips"`
	Nodes []Node `yaml:"nodes"`
}

// Config mirrors config.yml, the non-secret settings shared by core, workloads
// and the justfile.
type Config struct {
	Domain           string                           `yaml:"domain"`
	Cluster          ClusterConfig                    `yaml:"cluster"`
	Gateway          struct{ Name, Namespace string } `yaml:"gateway"`
	OpenBaoAddr      string                           `yaml:"openbao_addr"`
	ManifestsBranch  string                           `yaml:"manifests_branch"`
	ArgoCDManualSync bool                             `yaml:"argocd_manual_sync"`
	GrafanaAdmins    []string                         `yaml:"grafana_admins"`
	ACMEEmail        string                           `yaml:"acme_email"`
	Cloudflare       struct {
		ZoneID string `yaml:"zone_id"`
	} `yaml:"cloudflare"`
	Proxmox struct {
		Username string `yaml:"username"`
		Endpoint string `yaml:"endpoint"`
		NodeName string `yaml:"nodename"`
		IP       string `yaml:"ip"`
		Host     string `yaml:"host"`
	} `yaml:"proxmox"`
	Hetzner struct {
		ServerName string `yaml:"server_name"`
		Image      string `yaml:"image"`
		ServerType string `yaml:"server_type"`
		Location   string `yaml:"location"`
		SSHKey     string `yaml:"ssh_key"`
		VPSIP      string `yaml:"vps_ip"`
	} `yaml:"hetzner"`
}

// Host returns "<sub>.<domain>".
func (c Config) Host(sub string) string { return sub + "." + c.Domain }

// Chart is a pinned Helm chart source.
type Chart struct{ Repo, Version string }

// Image is a pinned container image.
type Image struct{ Image, Tag string }

// Ref returns the image reference as "<image>:<tag>".
func (i Image) Ref() string { return i.Image + ":" + i.Tag }

// Versions mirrors versions.yaml, the version pins for Talos, Kubernetes,
// charts and images.
type Versions struct {
	Talos struct {
		Version       string `yaml:"version"`
		BaseSchematic string `yaml:"base_schematic"`
		GPUSchematic  string `yaml:"gpu_schematic"`
	} `yaml:"talos"`
	Kubernetes struct {
		Version string `yaml:"version"`
	} `yaml:"kubernetes"`
	Charts map[string]Chart `yaml:"charts"`
	Images map[string]Image `yaml:"images"`
}

// C and V hold the embedded config.yml and versions.yaml, decoded at package
// init; a malformed file panics.
var (
	C Config
	V Versions
)

func init() {
	if err := yaml.Unmarshal(configYAML, &C); err != nil {
		panic(fmt.Errorf("config.yml: %w", err))
	}
	if err := yaml.Unmarshal(versionsYAML, &V); err != nil {
		panic(fmt.Errorf("versions.yaml: %w", err))
	}
}

// HelmChart returns the pinned chart called name. It panics on an unknown name
// so a typo fails the build instead of rendering an empty version.
func HelmChart(name string) Chart {
	c, ok := V.Charts[name]
	if !ok {
		panic("versions.yaml: no chart " + name)
	}
	return c
}

// ContainerImage returns the pinned image called name, panicking like
// [HelmChart] on an unknown name.
func ContainerImage(name string) Image {
	i, ok := V.Images[name]
	if !ok {
		panic("versions.yaml: no image " + name)
	}
	return i
}
