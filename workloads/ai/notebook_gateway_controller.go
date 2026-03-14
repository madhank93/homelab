package ai

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

// NewNotebookGatewayControllerChart creates a small controller that watches Notebook CRs
// and creates an HTTPRoute + ReferenceGrant for each one so the notebook URL
// /notebook/<namespace>/<name>/ is routable through the Cilium Gateway.
//
// Why this is needed:
//   The notebook-controller has USE_ISTIO=true and creates Istio VirtualServices per notebook.
//   Without Istio, those VirtualServices are no-ops. Gateway API has no dynamic path routing,
//   so we need a controller to create an HTTPRoute per Notebook in the kubeflow namespace
//   and a ReferenceGrant in the user namespace (required for cross-namespace service refs).
func NewNotebookGatewayControllerChart(scope constructs.Construct, id string, namespace string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String(namespace),
	})

	// ServiceAccount
	cdk8s.NewApiObject(chart, jsii.String("ngc-sa"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("ServiceAccount"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String("notebook-gateway-controller"), Namespace: jsii.String(namespace)},
	})

	// ClusterRole — needs to watch Notebooks (cluster-scoped read), manage HTTPRoutes in kubeflow ns,
	// manage ReferenceGrants in user namespaces.
	cdk8s.NewApiObject(chart, jsii.String("ngc-cr"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("rbac.authorization.k8s.io/v1"),
		Kind:       jsii.String("ClusterRole"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String("notebook-gateway-controller")},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/rules"), []map[string]any{
		{
			"apiGroups": []string{"kubeflow.org"},
			"resources": []string{"notebooks"},
			"verbs":     []string{"get", "list", "watch"},
		},
		{
			"apiGroups": []string{"gateway.networking.k8s.io"},
			"resources": []string{"httproutes"},
			"verbs":     []string{"get", "list", "create", "update", "patch", "delete"},
		},
		{
			"apiGroups": []string{"gateway.networking.k8s.io"},
			"resources": []string{"referencegrants"},
			"verbs":     []string{"get", "list", "create", "update", "patch", "delete"},
		},
	}))

	// ClusterRoleBinding
	crb := cdk8s.NewApiObject(chart, jsii.String("ngc-crb"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("rbac.authorization.k8s.io/v1"),
		Kind:       jsii.String("ClusterRoleBinding"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String("notebook-gateway-controller")},
	})
	crb.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/roleRef"), map[string]any{
		"apiGroup": "rbac.authorization.k8s.io",
		"kind":     "ClusterRole",
		"name":     "notebook-gateway-controller",
	}))
	crb.AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/subjects"), []map[string]any{
		{"kind": "ServiceAccount", "name": "notebook-gateway-controller", "namespace": namespace},
	}))

	// Deployment — Python controller using the in-cluster Kubernetes client
	controllerScript := `
import os, time, logging
from kubernetes import client, config, watch

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger("ngc")

config.load_incluster_config()
custom = client.CustomObjectsApi()
net_v1 = client.CustomObjectsApi()
core_v1 = client.CoreV1Api()

GATEWAY_NS   = "kube-system"
GATEWAY_NAME = "homelab-gateway"
ROUTE_NS     = "kubeflow"
HOSTNAME     = "kubeflow.madhan.app"
USERID_HEADER = "kubeflow-userid"
USERID_VALUE  = "user@example.com"

def route_name(ns, name):
    return f"notebook-{ns}-{name}"

def desired_httproute(ns, name):
    path = f"/notebook/{ns}/{name}/"
    return {
        "apiVersion": "gateway.networking.k8s.io/v1",
        "kind": "HTTPRoute",
        "metadata": {
            "name": route_name(ns, name),
            "namespace": ROUTE_NS,
            "labels": {"app.kubernetes.io/managed-by": "notebook-gateway-controller"},
        },
        "spec": {
            "parentRefs": [{"name": GATEWAY_NAME, "namespace": GATEWAY_NS}],
            "hostnames": [HOSTNAME],
            "rules": [{
                "matches": [{"path": {"type": "PathPrefix", "value": path}}],
                "filters": [{"type": "RequestHeaderModifier", "requestHeaderModifier": {
                    "set": [{"name": USERID_HEADER, "value": USERID_VALUE}]
                }}],
                "backendRefs": [{"name": name, "namespace": ns, "port": 80, "group": "", "kind": "Service"}],
            }],
        },
    }

def desired_refgrant(ns):
    return {
        "apiVersion": "gateway.networking.k8s.io/v1beta1",
        "kind": "ReferenceGrant",
        "metadata": {
            "name": "allow-kubeflow-gateway",
            "namespace": ns,
            "labels": {"app.kubernetes.io/managed-by": "notebook-gateway-controller"},
        },
        "spec": {
            "from": [{"group": "gateway.networking.k8s.io", "kind": "HTTPRoute", "namespace": ROUTE_NS}],
            "to":   [{"group": "", "kind": "Service"}],
        },
    }

def apply_route(ns, name):
    body = desired_httproute(ns, name)
    rname = route_name(ns, name)
    try:
        net_v1.get_namespaced_custom_object("gateway.networking.k8s.io", "v1", ROUTE_NS, "httproutes", rname)
        net_v1.patch_namespaced_custom_object("gateway.networking.k8s.io", "v1", ROUTE_NS, "httproutes", rname, body)
        log.info(f"Updated HTTPRoute {rname}")
    except client.exceptions.ApiException as e:
        if e.status == 404:
            net_v1.create_namespaced_custom_object("gateway.networking.k8s.io", "v1", ROUTE_NS, "httproutes", body)
            log.info(f"Created HTTPRoute {rname}")
        else:
            raise

def apply_refgrant(ns):
    body = desired_refgrant(ns)
    try:
        net_v1.get_namespaced_custom_object("gateway.networking.k8s.io", "v1beta1", ns, "referencegrants", "allow-kubeflow-gateway")
        net_v1.patch_namespaced_custom_object("gateway.networking.k8s.io", "v1beta1", ns, "referencegrants", "allow-kubeflow-gateway", body)
    except client.exceptions.ApiException as e:
        if e.status == 404:
            net_v1.create_namespaced_custom_object("gateway.networking.k8s.io", "v1beta1", ns, "referencegrants", body)
            log.info(f"Created ReferenceGrant in {ns}")
        else:
            raise

def delete_route(ns, name):
    rname = route_name(ns, name)
    try:
        net_v1.delete_namespaced_custom_object("gateway.networking.k8s.io", "v1", ROUTE_NS, "httproutes", rname)
        log.info(f"Deleted HTTPRoute {rname}")
    except client.exceptions.ApiException as e:
        if e.status != 404:
            raise

def sync_all():
    """Reconcile all existing notebooks on startup."""
    notebooks = custom.list_cluster_custom_object("kubeflow.org", "v1", "notebooks")
    for nb in notebooks.get("items", []):
        ns   = nb["metadata"]["namespace"]
        name = nb["metadata"]["name"]
        apply_refgrant(ns)
        apply_route(ns, name)
    log.info(f"Synced {len(notebooks.get('items',[]))} notebooks")

def run():
    sync_all()
    w = watch.Watch()
    log.info("Watching Notebook CRs...")
    while True:
        try:
            for event in w.stream(custom.list_cluster_custom_object, "kubeflow.org", "v1", "notebooks", timeout_seconds=300):
                etype = event["type"]
                nb    = event["object"]
                ns    = nb["metadata"]["namespace"]
                name  = nb["metadata"]["name"]
                if etype in ("ADDED", "MODIFIED"):
                    apply_refgrant(ns)
                    apply_route(ns, name)
                elif etype == "DELETED":
                    delete_route(ns, name)
        except Exception as exc:
            log.warning(f"Watch error: {exc}. Resyncing in 10s...")
            time.sleep(10)
            sync_all()

if __name__ == "__main__":
    run()
`

	cdk8s.NewApiObject(chart, jsii.String("ngc-deploy"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("apps/v1"),
		Kind:       jsii.String("Deployment"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String("notebook-gateway-controller"), Namespace: jsii.String(namespace)},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/spec"), map[string]any{
		"replicas": 1,
		"selector": map[string]any{"matchLabels": map[string]any{"app": "notebook-gateway-controller"}},
		"template": map[string]any{
			"metadata": map[string]any{"labels": map[string]any{"app": "notebook-gateway-controller"}},
			"spec": map[string]any{
				"serviceAccountName": "notebook-gateway-controller",
				"containers": []map[string]any{
					{
						"name":  "controller",
						"image": "python:3.11-slim",
						"command": []string{
							"sh", "-c",
							"pip install kubernetes --quiet --no-cache-dir && python /app/controller.py",
						},
						"env": []map[string]any{
							{"name": "HOME", "value": "/tmp"},
						},
						"securityContext": map[string]any{
							"allowPrivilegeEscalation": false,
							"runAsNonRoot":             true,
							"runAsUser":                int64(65534),
							"seccompProfile":           map[string]any{"type": "RuntimeDefault"},
							"capabilities":             map[string]any{"drop": []string{"ALL"}},
						},
						"volumeMounts": []map[string]any{
							{"name": "script", "mountPath": "/app"},
						},
						"resources": map[string]any{
							"requests": map[string]any{"cpu": "10m", "memory": "64Mi"},
							"limits":   map[string]any{"cpu": "100m", "memory": "128Mi"},
						},
					},
				},
				"volumes": []map[string]any{
					{
						"name": "script",
						"configMap": map[string]any{"name": "notebook-gateway-controller-script"},
					},
				},
			},
		},
	}))

	// ConfigMap holding the controller Python script
	cdk8s.NewApiObject(chart, jsii.String("ngc-cm"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("ConfigMap"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String("notebook-gateway-controller-script"), Namespace: jsii.String(namespace)},
	}).AddJsonPatch(cdk8s.JsonPatch_Add(jsii.String("/data"), map[string]any{
		"controller.py": controllerScript,
	}))

	return chart
}
