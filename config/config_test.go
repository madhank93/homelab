package config

import (
	"io/fs"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestConfigRequiredFields(t *testing.T) {
	for name, v := range map[string]string{
		"domain":             C.Domain,
		"cluster.name":       C.Cluster.Name,
		"cluster.subnet":     C.Cluster.Subnet,
		"gateway.name":       C.Gateway.Name,
		"gateway.namespace":  C.Gateway.Namespace,
		"openbao_addr":       C.OpenBaoAddr,
		"manifests_branch":   C.ManifestsBranch,
		"proxmox.endpoint":   C.Proxmox.Endpoint,
		"proxmox.nodename":   C.Proxmox.NodeName,
		"hetzner.vps_ip":     C.Hetzner.VPSIP,
		"cloudflare.zone_id": C.Cloudflare.ZoneID,
	} {
		if v == "" {
			t.Errorf("config.yml: %s is empty", name)
		}
	}
}

func TestClusterAddresses(t *testing.T) {
	_, subnet, err := net.ParseCIDR(C.Cluster.Subnet)
	if err != nil {
		t.Fatalf("config.yml: cluster.subnet %q: %v", C.Cluster.Subnet, err)
	}
	inSubnet := map[string]string{
		"cluster.vip":           C.Cluster.VIP,
		"cluster.router":        C.Cluster.Router,
		"cluster.gateway_ip":    C.Cluster.GatewayIP,
		"cluster.lb_pool.start": C.Cluster.LBPool.Start,
		"cluster.lb_pool.stop":  C.Cluster.LBPool.Stop,
	}
	for _, n := range C.Cluster.Nodes {
		inSubnet["node "+n.Name] = n.IP
	}
	for name, ip := range inSubnet {
		if parsed := net.ParseIP(ip); parsed == nil || !subnet.Contains(parsed) {
			t.Errorf("config.yml: %s %q is not an IP in %s", name, ip, C.Cluster.Subnet)
		}
	}
}

func TestNodes(t *testing.T) {
	if len(C.Cluster.Nodes) == 0 {
		t.Fatal("config.yml: no cluster.nodes")
	}
	names, ips := map[string]bool{}, map[string]bool{}
	roles := map[string]int{}
	for _, n := range C.Cluster.Nodes {
		if n.Name == "" || n.IP == "" || n.Cores <= 0 || n.Memory <= 0 || n.Disk <= 0 {
			t.Errorf("config.yml: node %+v has empty name/ip or non-positive resources", n)
		}
		if n.Role != "control" && n.Role != "worker" {
			t.Errorf("config.yml: node %s role %q, want control or worker", n.Name, n.Role)
		}
		if names[n.Name] || ips[n.IP] {
			t.Errorf("config.yml: node %s (%s) duplicates a name or IP", n.Name, n.IP)
		}
		names[n.Name], ips[n.IP] = true, true
		roles[n.Role]++
	}
	if roles["control"]%2 == 0 {
		t.Errorf("config.yml: %d control-plane nodes; etcd needs an odd count", roles["control"])
	}
}

func TestFixedLBIPs(t *testing.T) {
	start, err1 := netip.ParseAddr(C.Cluster.LBPool.Start)
	stop, err2 := netip.ParseAddr(C.Cluster.LBPool.Stop)
	if err1 != nil || err2 != nil {
		t.Fatalf("config.yml: cluster.lb_pool %+v is not a pair of IPs", C.Cluster.LBPool)
	}
	// Announcing a node's or the VIP's address over L2 takes it off the LAN.
	taken := map[string]string{C.Cluster.VIP: "cluster.vip", C.Cluster.GatewayIP: "cluster.gateway_ip"}
	for _, n := range C.Cluster.Nodes {
		taken[n.IP] = "node " + n.Name
	}
	for name, ip := range map[string]string{
		"cs2":     C.Cluster.LBIPs.CS2,
		"ollama":  C.Cluster.LBIPs.Ollama,
		"adguard": C.Cluster.LBIPs.AdGuard,
	} {
		addr, err := netip.ParseAddr(ip)
		if err != nil || addr.Less(start) || stop.Less(addr) {
			t.Errorf("config.yml: cluster.lb_ips.%s %q is not in lb_pool %s-%s", name, ip, start, stop)
		}
		if owner, ok := taken[ip]; ok {
			t.Errorf("config.yml: cluster.lb_ips.%s %s is already used by %s", name, ip, owner)
		}
		taken[ip] = "cluster.lb_ips." + name
	}
}

func TestHost(t *testing.T) {
	c := Config{Domain: "example.com"}
	for sub, want := range map[string]string{"grafana": "grafana.example.com", "a.b": "a.b.example.com"} {
		if got := c.Host(sub); got != want {
			t.Errorf("Host(%q) = %q, want %q", sub, got, want)
		}
	}
}

func TestNodeHasGPU(t *testing.T) {
	for _, tc := range []struct {
		node Node
		want bool
	}{
		{Node{}, false},
		{Node{PcieIDs: []string{"0000:09:00.0"}}, true},
	} {
		if got := tc.node.HasGPU(); got != tc.want {
			t.Errorf("%+v.HasGPU() = %v, want %v", tc.node, got, tc.want)
		}
	}
}

func TestImageRef(t *testing.T) {
	if got := (Image{Image: "registry.k8s.io/pause", Tag: "3.10"}).Ref(); got != "registry.k8s.io/pause:3.10" {
		t.Errorf("Ref() = %q", got)
	}
}

func TestVersionsRequiredFields(t *testing.T) {
	for name, v := range map[string]string{
		"talos.version":        V.Talos.Version,
		"talos.base_schematic": V.Talos.BaseSchematic,
		"talos.gpu_schematic":  V.Talos.GPUSchematic,
		"kubernetes.version":   V.Kubernetes.Version,
	} {
		if v == "" {
			t.Errorf("versions.yaml: %s is empty", name)
		}
	}
	for name, c := range V.Charts {
		if c.Repo == "" || c.Version == "" {
			t.Errorf("versions.yaml: chart %s has empty repo or version", name)
		}
	}
	for name, i := range V.Images {
		if i.Image == "" || i.Tag == "" || i.Tag == "latest" {
			t.Errorf("versions.yaml: image %s needs an image and a non-latest tag, got %q", name, i.Ref())
		}
	}
}

func TestLookupsPanicOnUnknownName(t *testing.T) {
	for name, fn := range map[string]func(){
		"HelmChart":      func() { HelmChart("no-such-chart") },
		"ContainerImage": func() { ContainerImage("no-such-image") },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("want panic for unknown name")
				}
			}()
			fn()
		})
	}
}

// Every name the Go code looks up must exist in versions.yaml, since lookups
// panic at synth/deploy time. Scanning source also covers core/, which cannot
// build without generated SDKs.
func TestReferencedPinsExist(t *testing.T) {
	re := regexp.MustCompile(`config\.(HelmChart|ContainerImage)\("([^"]+)"\)`)
	seen := 0
	for _, dir := range []string{"../workloads", "../core"} {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && (d.Name() == "sdks" || d.Name() == "imports") {
				return filepath.SkipDir
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, m := range re.FindAllStringSubmatch(string(src), -1) {
				seen++
				var ok bool
				if m[1] == "HelmChart" {
					_, ok = V.Charts[m[2]]
				} else {
					_, ok = V.Images[m[2]]
				}
				if !ok {
					t.Errorf("%s: %s(%q) has no entry in versions.yaml", path, m[1], m[2])
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if seen == 0 {
		t.Error("found no config.HelmChart/ContainerImage calls; the scan is likely broken")
	}
}
