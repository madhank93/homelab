package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"go.yaml.in/yaml/v3"
)

// Policy tests over the synthesized output of every entry in apps. Synthesis
// pulls Helm charts over the network, so these skip under -short.

type object struct {
	file string // relative to the synth root, as under app/
	m    map[string]any
}

func (o object) kind() string { return str(o.m, "kind") }
func (o object) name() string { return str(o.m, "metadata", "name") }
func (o object) namespace() string {
	return str(o.m, "metadata", "namespace")
}

var (
	synthOnce sync.Once
	synthObjs map[string][]object // app dir -> objects
	synthErr  error
)

// synthAll renders every app once per test binary into a temp dir.
func synthAll(t *testing.T) map[string][]object {
	t.Helper()
	if testing.Short() {
		t.Skip("synthesis needs helm and network access")
	}
	synthOnce.Do(func() {
		root, err := os.MkdirTemp("", "workloads-synth-")
		if err != nil {
			synthErr = err
			return
		}
		defer os.RemoveAll(root)
		for _, entry := range apps {
			app := cdk8s.NewApp(&cdk8s.AppProps{
				Outdir:         jsii.String(filepath.Join(root, entry.dir)),
				YamlOutputType: cdk8s.YamlOutputType_FILE_PER_RESOURCE,
			})
			entry.build(app)
			app.Synth()
		}
		synthObjs, synthErr = loadObjects(root)
	})
	if synthErr != nil {
		t.Fatalf("synthesize apps: %v", synthErr)
	}
	return synthObjs
}

func loadObjects(root string) (map[string][]object, error) {
	out := map[string][]object{}
	for _, entry := range apps {
		out[entry.dir] = nil
	}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		dec := yaml.NewDecoder(bytes.NewReader(raw))
		for {
			var m map[string]any
			if err := dec.Decode(&m); errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return errors.New(rel + ": " + err.Error())
			}
			if m != nil {
				dir, _, _ := strings.Cut(rel, "/")
				out[dir] = append(out[dir], object{file: rel, m: m})
			}
		}
		return nil
	})
	return out, err
}

func allObjects(t *testing.T) []object {
	var all []object
	for _, objs := range synthAll(t) {
		all = append(all, objs...)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].file < all[j].file })
	return all
}

// str walks nested maps by key and returns the string at the end, or "".
func str(m map[string]any, keys ...string) string {
	var cur any = m
	for _, k := range keys {
		mm, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = mm[k]
	}
	s, _ := cur.(string)
	return s
}

func list(v any) []map[string]any {
	items, _ := v.([]any)
	var out []map[string]any
	for _, it := range items {
		if m, ok := it.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

// podSpecs finds every map holding a containers list, which covers Pods,
// workload templates and CRs that embed a pod spec.
func podSpecs(v any) []map[string]any {
	var out []map[string]any
	switch x := v.(type) {
	case map[string]any:
		if _, ok := x["containers"].([]any); ok {
			out = append(out, x)
		}
		for _, child := range x {
			out = append(out, podSpecs(child)...)
		}
	case []any:
		for _, child := range x {
			out = append(out, podSpecs(child)...)
		}
	}
	return out
}

func TestEveryAppRendersObjects(t *testing.T) {
	objs := synthAll(t)
	for _, entry := range apps {
		if len(objs[entry.dir]) == 0 {
			t.Errorf("app %q rendered no objects", entry.dir)
		}
	}
}

// readAllowlist mirrors scripts/manifest-secret-guard.sh's parsing.
func readAllowlist(t *testing.T) map[string]bool {
	t.Helper()
	f, err := os.Open("../scripts/manifest-secret-allowlist.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	allowed := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line, _, _ := strings.Cut(sc.Text(), "#")
		if line = strings.TrimSpace(line); line != "" {
			allowed[line] = true
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return allowed
}

// The manifests branch is public, so a Secret with a payload is a leak unless
// allowlisted; the allowlist may only shrink (see scripts/manifest-secret-guard.sh).
func TestNoSecretPayloads(t *testing.T) {
	allowed := readAllowlist(t)
	leaking := map[string]bool{}
	for _, o := range allObjects(t) {
		if o.kind() != "Secret" {
			continue
		}
		data, _ := o.m["data"].(map[string]any)
		stringData, _ := o.m["stringData"].(map[string]any)
		if len(data) == 0 && len(stringData) == 0 {
			continue
		}
		leaking[o.file] = true
		if !allowed[o.file] {
			t.Errorf("%s: Secret %s/%s has data/stringData and would be published; source it from OpenBao",
				o.file, o.namespace(), o.name())
		}
	}
	for entry := range allowed {
		// Rendered by kustomize in CI, not by synthesis.
		if entry == "kubeflow/kubeflow-manifests.yaml" {
			continue
		}
		if !leaking[entry] {
			t.Errorf("%s is allowlisted but renders no Secret payload; remove it from scripts/manifest-secret-allowlist.txt", entry)
		}
	}
}

// pinned reports whether ref carries a digest or a non-latest tag.
func pinned(ref string) bool {
	if strings.Contains(ref, "@sha256:") {
		return true
	}
	last := ref[strings.LastIndex(ref, "/")+1:]
	_, tag, ok := strings.Cut(last, ":")
	return ok && tag != "" && tag != "latest"
}

// Scope: images rendered as container specs; images inside CR fields like
// {repository, tag} are not checked. Argo CD ignores Helm test hooks, so
// those Pods are skipped.
func TestImagesPinned(t *testing.T) {
	for _, o := range allObjects(t) {
		if strings.Contains(str(o.m, "metadata", "annotations", "helm.sh/hook"), "test") {
			continue
		}
		for _, spec := range podSpecs(o.m) {
			for _, key := range []string{"initContainers", "containers"} {
				for _, c := range list(spec[key]) {
					if img := str(c, "image"); !pinned(img) {
						t.Errorf("%s: %s %s container %q image %q has no tag/digest or uses :latest",
							o.file, o.kind(), o.name(), str(c, "name"), img)
					}
				}
			}
		}
	}
}

func TestPinned(t *testing.T) {
	for ref, want := range map[string]bool{
		"alpine":                         false,
		"alpine:latest":                  false,
		"alpine:3.20":                    true,
		"localhost:5000/app":             false,
		"localhost:5000/app:v1":          true,
		"ghcr.io/x/y@sha256:abc":         true,
		"registry.k8s.io/pause:3.10":     true,
		"ghcr.io/kyverno/checker:latest": false,
	} {
		if got := pinned(ref); got != want {
			t.Errorf("pinned(%q) = %v, want %v", ref, got, want)
		}
	}
}

func TestHTTPRoutesUseSharedGateway(t *testing.T) {
	gw := config.C.Gateway
	found := 0
	for _, o := range allObjects(t) {
		if o.kind() != "HTTPRoute" {
			continue
		}
		found++
		spec, _ := o.m["spec"].(map[string]any)
		refs := list(spec["parentRefs"])
		if len(refs) == 0 {
			t.Errorf("%s: HTTPRoute %s has no parentRefs", o.file, o.name())
		}
		for _, ref := range refs {
			if str(ref, "name") != gw.Name || str(ref, "namespace") != gw.Namespace {
				t.Errorf("%s: HTTPRoute %s parentRef %s/%s, want Gateway %s/%s",
					o.file, o.name(), str(ref, "namespace"), str(ref, "name"), gw.Namespace, gw.Name)
			}
		}
	}
	if found == 0 {
		t.Error("no HTTPRoutes rendered; the walk is likely broken")
	}
}

// openBaoRoles parses `roles="app:sa ..."` from scripts/openbao-setup.sh into
// app -> ServiceAccount. Each role is bound to namespace = app.
func openBaoRoles(t *testing.T) map[string]string {
	t.Helper()
	raw, err := os.ReadFile("../scripts/openbao-setup.sh")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?m)^roles="([^"]+)"`).FindSubmatch(raw)
	if m == nil {
		t.Fatal(`scripts/openbao-setup.sh: no roles="..." line`)
	}
	roles := map[string]string{}
	for f := range strings.FieldsSeq(string(m[1])) {
		app, sa, ok := strings.Cut(f, ":")
		if !ok {
			t.Fatalf("scripts/openbao-setup.sh: malformed role %q", f)
		}
		roles[app] = sa
	}
	return roles
}

// Every SecretProviderClass must use a role that openbao-setup.sh creates, live
// in that role's namespace, and be mounted by a pod running as the role's
// ServiceAccount (the mount is what triggers the Secret mirror).
func TestSecretProviderClassRoles(t *testing.T) {
	roles := openBaoRoles(t)
	objs := allObjects(t)
	found := 0
	for _, spc := range objs {
		if spc.kind() != "SecretProviderClass" {
			continue
		}
		found++
		role := str(spc.m, "spec", "parameters", "roleName")
		sa, ok := roles[role]
		if !ok {
			t.Errorf("%s: SecretProviderClass %s roleName %q is not a role in scripts/openbao-setup.sh",
				spc.file, spc.name(), role)
			continue
		}
		if spc.namespace() != role {
			t.Errorf("%s: SecretProviderClass %s is in namespace %q, but role %q is bound to namespace %q",
				spc.file, spc.name(), spc.namespace(), role, role)
		}
		var mountedBy []string
		for _, o := range objs {
			if o.namespace() != spc.namespace() {
				continue
			}
			for _, ps := range podSpecs(o.m) {
				for _, v := range list(ps["volumes"]) {
					if str(v, "csi", "volumeAttributes", "secretProviderClass") != spc.name() {
						continue
					}
					got := str(ps, "serviceAccountName")
					if got == "" {
						got = "default"
					}
					mountedBy = append(mountedBy, o.file)
					if got != sa {
						t.Errorf("%s: %s mounts SecretProviderClass %s as ServiceAccount %q, but role %q binds %q",
							o.file, o.name(), spc.name(), got, role, sa)
					}
				}
			}
		}
		if len(mountedBy) == 0 {
			t.Errorf("%s: no pod mounts SecretProviderClass %s, so its Secret is never mirrored", spc.file, spc.name())
		}
	}
	if found == 0 {
		t.Error("no SecretProviderClasses rendered; the walk is likely broken")
	}
}
