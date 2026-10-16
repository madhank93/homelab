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

func (n Node) HasGPU() bool { return len(n.PcieIDs) > 0 }

type ClusterConfig struct {
	Name          string                       `yaml:"name"`
	VIP           string                       `yaml:"vip"`
	Subnet        string                       `yaml:"subnet"`
	Router        string                       `yaml:"router"`
	Nameservers   []string                     `yaml:"nameservers"`
	GatewayIP     string                       `yaml:"gateway_ip"`
	LBPool        struct{ Start, Stop string } `yaml:"lb_pool"`
	NetbirdCIDR   string                       `yaml:"netbird_cidr"`
	NetbirdRouter string                       `yaml:"netbird_router"`
	Nodes         []Node                       `yaml:"nodes"`
}

type Config struct {
	Domain          string                           `yaml:"domain"`
	Cluster         ClusterConfig                    `yaml:"cluster"`
	Gateway         struct{ Name, Namespace string } `yaml:"gateway"`
	OpenBaoAddr     string                           `yaml:"openbao_addr"`
	ManifestsBranch string                           `yaml:"manifests_branch"`
	Cloudflare      struct {
		ZoneID string `yaml:"zone_id"`
	} `yaml:"cloudflare"`
	Proxmox struct {
		Username string `yaml:"username"`
		Endpoint string `yaml:"endpoint"`
		NodeName string `yaml:"nodename"`
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

type Chart struct{ Repo, Version string }

type Image struct{ Image, Tag string }

func (i Image) Ref() string { return i.Image + ":" + i.Tag }

type Versions struct {
	Talos struct {
		Version       string `yaml:"version"`
		BaseSchematic string `yaml:"base_schematic"`
		GPUSchematic  string `yaml:"gpu_schematic"`
	} `yaml:"talos"`
	Charts map[string]Chart `yaml:"charts"`
	Images map[string]Image `yaml:"images"`
}

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

// HelmChart returns the pinned chart, failing loudly on a typo rather than
// rendering an empty version.
func HelmChart(name string) Chart {
	c, ok := V.Charts[name]
	if !ok {
		panic("versions.yaml: no chart " + name)
	}
	return c
}

// ContainerImage returns the pinned image; see HelmChart.
func ContainerImage(name string) Image {
	i, ok := V.Images[name]
	if !ok {
		panic("versions.yaml: no image " + name)
	}
	return i
}
