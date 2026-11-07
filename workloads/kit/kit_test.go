package kit

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
)

// render builds a fresh test chart with fn and returns its objects as plain
// JSON values, so assertions compare against map[string]any.
func render(t *testing.T, fn func(cdk8s.Chart)) []map[string]any {
	t.Helper()
	chart := cdk8s.Testing_Chart()
	fn(chart)
	raw, err := json.Marshal(chart.ToJson())
	if err != nil {
		t.Fatal(err)
	}
	var objs []map[string]any
	if err := json.Unmarshal(raw, &objs); err != nil {
		t.Fatal(err)
	}
	return objs
}

// only returns the single object of kind, failing otherwise.
func only(t *testing.T, objs []map[string]any, kind string) map[string]any {
	t.Helper()
	var found []map[string]any
	for _, o := range objs {
		if o["kind"] == kind {
			found = append(found, o)
		}
	}
	if len(found) != 1 {
		t.Fatalf("want 1 %s, got %d", kind, len(found))
	}
	return found[0]
}

// get walks nested maps and slices; ints index slices.
func get(v any, path ...any) any {
	for _, p := range path {
		switch k := p.(type) {
		case string:
			m, _ := v.(map[string]any)
			v = m[k]
		case int:
			s, _ := v.([]any)
			if k >= len(s) {
				return nil
			}
			v = s[k]
		}
	}
	return v
}

// normalize round-trips v through JSON, dropping null fields, so Go values
// compare with rendered output.
func normalize(t *testing.T, v any) any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return dropNulls(out)
}

func dropNulls(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			if child == nil {
				delete(x, k)
			} else {
				x[k] = dropNulls(child)
			}
		}
	case []any:
		for i, child := range x {
			x[i] = dropNulls(child)
		}
	}
	return v
}

func TestNamespace(t *testing.T) {
	for _, tc := range []struct {
		name       string
		privileged bool
		wantLabels map[string]any
	}{
		{"plain", false, nil},
		{"privileged", true, map[string]any{
			"pod-security.kubernetes.io/enforce": "privileged",
			"pod-security.kubernetes.io/audit":   "privileged",
			"pod-security.kubernetes.io/warn":    "privileged",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ns := only(t, render(t, func(c cdk8s.Chart) { Namespace(c, "demo", tc.privileged) }), "Namespace")
			if got := get(ns, "metadata", "name"); got != "demo" {
				t.Errorf("name = %v, want demo", got)
			}
			labels, _ := get(ns, "metadata", "labels").(map[string]any)
			if len(labels) == 0 {
				labels = nil
			}
			if !reflect.DeepEqual(labels, tc.wantLabels) {
				t.Errorf("labels = %v, want %v", labels, tc.wantLabels)
			}
		})
	}
}

func TestBackend(t *testing.T) {
	want := map[string]any{"group": "", "kind": "Service", "name": "web", "port": 8080, "weight": 1}
	if got := Backend("web", 8080); !reflect.DeepEqual(got, want) {
		t.Errorf("Backend = %v, want %v", got, want)
	}
}

func TestRoute(t *testing.T) {
	route := only(t, render(t, func(c cdk8s.Chart) { Route(c, "web", "app", "web-svc", 80) }), "HTTPRoute")
	gw := config.C.Gateway

	checks := []struct {
		name string
		path []any
		want any
	}{
		{"apiVersion", []any{"apiVersion"}, "gateway.networking.k8s.io/v1"},
		{"name", []any{"metadata", "name"}, "web"},
		{"hostname", []any{"spec", "hostnames", 0}, "app." + config.C.Domain},
		{"gateway name", []any{"spec", "parentRefs", 0, "name"}, gw.Name},
		{"gateway namespace", []any{"spec", "parentRefs", 0, "namespace"}, gw.Namespace},
		{"path prefix", []any{"spec", "rules", 0, "matches", 0, "path", "value"}, "/"},
		{"backend", []any{"spec", "rules", 0, "backendRefs", 0}, normalize(t, Backend("web-svc", 80))},
	}
	for _, c := range checks {
		if got := get(route, c.path...); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestRouteRulesKeepsCallerRules(t *testing.T) {
	rules := []map[string]any{{"filters": []map[string]any{{"type": "RequestRedirect"}}}}
	route := only(t, render(t, func(c cdk8s.Chart) { RouteRules(c, "r", "sub", rules) }), "HTTPRoute")
	if got, want := get(route, "spec", "rules"), normalize(t, rules); !reflect.DeepEqual(got, want) {
		t.Errorf("rules = %v, want %v", got, want)
	}
}

func TestSecretProviderClass(t *testing.T) {
	for _, tc := range []struct {
		name     string
		objects  []string
		sync     []Sync
		wantData any
	}{
		{"no sync", []string{"A"}, nil, []any{}},
		{"sync", []string{"A", "B"}, []Sync{{"A", "a"}, {"B", "b"}}, []any{
			map[string]any{"objectName": "A", "key": "a"},
			map[string]any{"objectName": "B", "key": "b"},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spc := only(t, render(t, func(c cdk8s.Chart) {
				SecretProviderClass(c, "app-secrets", "app", tc.objects, "app-mirror", tc.sync...)
			}), "SecretProviderClass")

			if got := get(spc, "spec", "provider"); got != "openbao" {
				t.Errorf("provider = %v", got)
			}
			if got := get(spc, "spec", "parameters", "roleName"); got != "app" {
				t.Errorf("roleName = %v, want app", got)
			}
			if got := get(spc, "spec", "parameters", "vaultAddress"); got != config.C.OpenBaoAddr {
				t.Errorf("vaultAddress = %v, want %v", got, config.C.OpenBaoAddr)
			}
			objs, _ := get(spc, "spec", "parameters", "objects").(string)
			for _, o := range tc.objects {
				if !strings.Contains(objs, `objectName: "`+o+`"`) || !strings.Contains(objs, `secretKey: "`+o+`"`) {
					t.Errorf("objects missing %s:\n%s", o, objs)
				}
			}
			if n := strings.Count(objs, `secretPath: "secret/data/app"`); n != len(tc.objects) {
				t.Errorf("objects has %d secretPath entries for role app, want %d:\n%s", n, len(tc.objects), objs)
			}
			if got := get(spc, "spec", "secretObjects", 0, "secretName"); got != "app-mirror" {
				t.Errorf("secretName = %v, want app-mirror", got)
			}
			if got := get(spc, "spec", "secretObjects", 0, "data"); !reflect.DeepEqual(got, tc.wantData) {
				t.Errorf("data = %v, want %v", got, tc.wantData)
			}
		})
	}
}

// The Helm-values forms must stay identical to the typed forms.
func TestSecretsValuesMatchTyped(t *testing.T) {
	for _, tc := range []struct {
		name         string
		typed, value any
	}{
		{"volume", SecretsVolume("spc"), SecretsVolumeValues("spc")},
		{"mount", SecretsMount(), SecretsMountValues()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := normalize(t, tc.typed), normalize(t, tc.value); !reflect.DeepEqual(got, want) {
				t.Errorf("typed = %v\nvalues = %v", got, want)
			}
		})
	}
}

func TestSecretSync(t *testing.T) {
	objs := render(t, func(c cdk8s.Chart) { SecretSync(c, "spc") })
	only(t, objs, "ServiceAccount")
	dep := only(t, objs, "Deployment")
	pod := get(dep, "spec", "template", "spec")

	if got := get(pod, "serviceAccountName"); got != "secret-sync" {
		t.Errorf("serviceAccountName = %v, want secret-sync", got)
	}
	if got := get(pod, "volumes", 0, "csi", "volumeAttributes", "secretProviderClass"); got != "spc" {
		t.Errorf("volume secretProviderClass = %v, want spc", got)
	}
	if got, want := get(pod, "containers", 0, "image"), config.ContainerImage("pause").Ref(); got != want {
		t.Errorf("image = %v, want %v", got, want)
	}
	if got := get(pod, "containers", 0, "volumeMounts", 0, "mountPath"); got != "/mnt/secrets" {
		t.Errorf("mountPath = %v, want /mnt/secrets", got)
	}
	if got := get(dep, "spec", "template", "metadata", "annotations", "reloader.stakater.com/auto"); got != "true" {
		t.Errorf("reloader annotation = %v, want true", got)
	}
}
