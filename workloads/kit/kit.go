// Package kit holds the building blocks every workload chart repeats: its
// namespace, its Gateway route, and OpenBao secrets delivered through the
// Secrets Store CSI driver.
package kit

import (
	"fmt"
	"strings"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/workloads/imports/k8s"
)

// Namespace creates the namespace. Privileged sets the Pod Security level for
// workloads that need host access (CNI-adjacent agents, storage, eBPF).
func Namespace(scope constructs.Construct, name string, privileged bool) {
	var labels *map[string]*string
	if privileged {
		labels = &map[string]*string{
			"pod-security.kubernetes.io/enforce": jsii.String("privileged"),
			"pod-security.kubernetes.io/audit":   jsii.String("privileged"),
			"pod-security.kubernetes.io/warn":    jsii.String("privileged"),
		}
	}
	k8s.NewKubeNamespace(scope, jsii.String("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String(name), Labels: labels},
	})
}

// Route exposes svc:port at <sub>.<domain> through the shared Gateway.
func Route(scope constructs.Construct, name, sub, svc string, port int) {
	RouteRules(scope, name, sub, []map[string]any{{
		"matches":     []map[string]any{{"path": map[string]any{"type": "PathPrefix", "value": "/"}}},
		"backendRefs": []map[string]any{Backend(svc, port)},
	}})
}

// RouteRules is Route with caller-supplied rules (redirects, header filters).
func RouteRules(scope constructs.Construct, name, sub string, rules []map[string]any) {
	gw := config.C.Gateway
	cdk8s.NewApiObject(scope, jsii.String(name+"-httproute"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("gateway.networking.k8s.io/v1"),
		Kind:       jsii.String("HTTPRoute"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String(name)},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"parentRefs": []map[string]any{
			{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": gw.Name, "namespace": gw.Namespace},
		},
		"hostnames": []string{config.C.Host(sub)},
		"rules":     rules,
	}))
}

// Backend is one HTTPRoute backendRef to a Service.
func Backend(svc string, port int) map[string]any {
	return map[string]any{"group": "", "kind": "Service", "name": svc, "port": port, "weight": 1}
}

// Sync maps an OpenBao key (Object) to a key in the mirrored Secret.
type Sync struct{ Object, Key string }

// SecretProviderClass reads objects from OpenBao secret/<role> with the
// Kubernetes auth role of the same name, and mirrors the synced keys into the
// Secret secretName. The mirror only happens while a pod mounts the CSI
// volume (SecretsVolume).
func SecretProviderClass(scope constructs.Construct, name, role string, objects []string, secretName string, sync ...Sync) {
	var objs []string
	for _, o := range objects {
		objs = append(objs, fmt.Sprintf("- objectName: %q\n  secretPath: \"secret/data/%s\"\n  secretKey: %q", o, role, o))
	}
	var data []map[string]any
	for _, s := range sync {
		data = append(data, map[string]any{"objectName": s.Object, "key": s.Key})
	}
	cdk8s.NewApiObject(scope, jsii.String(name+"-spc"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("secrets-store.csi.x-k8s.io/v1"),
		Kind:       jsii.String("SecretProviderClass"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String(name)},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"provider": "openbao",
		"parameters": map[string]any{
			"vaultAddress": config.C.OpenBaoAddr,
			"roleName":     role,
			"objects":      strings.Join(objs, "\n"),
		},
		"secretObjects": []map[string]any{
			{"secretName": secretName, "type": "Opaque", "data": data},
		},
	}))
}

const secretsVolume = "openbao-secrets"

// SecretsVolume is the CSI volume for the SecretProviderClass spc.
func SecretsVolume(spc string) *k8s.Volume {
	return &k8s.Volume{
		Name: jsii.String(secretsVolume),
		Csi: &k8s.CsiVolumeSource{
			Driver:           jsii.String("secrets-store.csi.k8s.io"),
			ReadOnly:         jsii.Bool(true),
			VolumeAttributes: &map[string]*string{"secretProviderClass": jsii.String(spc)},
		},
	}
}

// SecretsMount mounts SecretsVolume at /mnt/secrets.
func SecretsMount() *k8s.VolumeMount {
	return &k8s.VolumeMount{Name: jsii.String(secretsVolume), MountPath: jsii.String("/mnt/secrets"), ReadOnly: jsii.Bool(true)}
}

// SecretsVolumeValues and SecretsMountValues are the Helm-values forms of
// SecretsVolume and SecretsMount, for charts that accept extra volumes.
func SecretsVolumeValues(spc string) map[string]any {
	return map[string]any{
		"name": secretsVolume,
		"csi": map[string]any{
			"driver":           "secrets-store.csi.k8s.io",
			"readOnly":         true,
			"volumeAttributes": map[string]any{"secretProviderClass": spc},
		},
	}
}

func SecretsMountValues() map[string]any {
	return map[string]any{"name": secretsVolume, "mountPath": "/mnt/secrets", "readOnly": true}
}

// SecretSync runs a pause pod that only mounts SecretsVolume, for charts that
// cannot mount extra volumes themselves; the mount triggers the Secret mirror.
func SecretSync(scope constructs.Construct, spc string) {
	k8s.NewKubeServiceAccount(scope, jsii.String("secret-sync-sa"), &k8s.KubeServiceAccountProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("secret-sync")},
	})
	labels := &map[string]*string{"app": jsii.String("secret-sync")}
	k8s.NewKubeDeployment(scope, jsii.String("secret-sync"), &k8s.KubeDeploymentProps{
		Metadata: &k8s.ObjectMeta{Name: jsii.String("secret-sync")},
		Spec: &k8s.DeploymentSpec{
			Replicas: jsii.Number(1),
			Selector: &k8s.LabelSelector{MatchLabels: labels},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{
					Labels:      labels,
					Annotations: &map[string]*string{"reloader.stakater.com/auto": jsii.String("true")},
				},
				Spec: &k8s.PodSpec{
					ServiceAccountName: jsii.String("secret-sync"),
					Volumes:            &[]*k8s.Volume{SecretsVolume(spc)},
					Containers: &[]*k8s.Container{{
						Name:         jsii.String("pause"),
						Image:        jsii.String(config.ContainerImage("pause").Ref()),
						VolumeMounts: &[]*k8s.VolumeMount{SecretsMount()},
					}},
				},
			},
		},
	})
}
