// headlamp
package headlamp

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterClass(
		"headlamp.Headlamp",
		reflect.TypeOf((*Headlamp)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "helm", GoGetter: "Helm"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "with", GoMethod: "With"},
		},
		func() interface{} {
			j := jsiiProxy_Headlamp{}
			_jsii_.InitJsiiProxy(&j.Type__constructsConstruct)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampClusterRoleBinding",
		reflect.TypeOf((*HeadlampClusterRoleBinding)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampConfig",
		reflect.TypeOf((*HeadlampConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampConfigClusterInventory",
		reflect.TypeOf((*HeadlampConfigClusterInventory)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampConfigClusterInventoryPlugins",
		reflect.TypeOf((*HeadlampConfigClusterInventoryPlugins)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampConfigOidc",
		reflect.TypeOf((*HeadlampConfigOidc)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampConfigOidcExternalSecret",
		reflect.TypeOf((*HeadlampConfigOidcExternalSecret)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampConfigOidcSecret",
		reflect.TypeOf((*HeadlampConfigOidcSecret)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampConfigStaticPlugins",
		reflect.TypeOf((*HeadlampConfigStaticPlugins)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampEnv",
		reflect.TypeOf((*HeadlampEnv)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampEnvValueFrom",
		reflect.TypeOf((*HeadlampEnvValueFrom)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampEnvValueFromConfigMapKeyRef",
		reflect.TypeOf((*HeadlampEnvValueFromConfigMapKeyRef)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampEnvValueFromFieldRef",
		reflect.TypeOf((*HeadlampEnvValueFromFieldRef)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampEnvValueFromResourceFieldRef",
		reflect.TypeOf((*HeadlampEnvValueFromResourceFieldRef)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampEnvValueFromSecretKeyRef",
		reflect.TypeOf((*HeadlampEnvValueFromSecretKeyRef)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampHostAliases",
		reflect.TypeOf((*HeadlampHostAliases)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampHttpRoute",
		reflect.TypeOf((*HeadlampHttpRoute)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampHttpRouteParentRefs",
		reflect.TypeOf((*HeadlampHttpRouteParentRefs)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampImage",
		reflect.TypeOf((*HeadlampImage)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"headlamp.HeadlampImagePullPolicy",
		reflect.TypeOf((*HeadlampImagePullPolicy)(nil)).Elem(),
		map[string]interface{}{
			"ALWAYS": HeadlampImagePullPolicy_ALWAYS,
			"IF_NOT_PRESENT": HeadlampImagePullPolicy_IF_NOT_PRESENT,
			"NEVER": HeadlampImagePullPolicy_NEVER,
		},
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampImagePullSecrets",
		reflect.TypeOf((*HeadlampImagePullSecrets)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampIngress",
		reflect.TypeOf((*HeadlampIngress)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampIngressHosts",
		reflect.TypeOf((*HeadlampIngressHosts)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampIngressHostsPaths",
		reflect.TypeOf((*HeadlampIngressHostsPaths)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampIngressHostsPathsBackend",
		reflect.TypeOf((*HeadlampIngressHostsPathsBackend)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampIngressHostsPathsBackendService",
		reflect.TypeOf((*HeadlampIngressHostsPathsBackendService)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampIngressHostsPathsBackendServicePort",
		reflect.TypeOf((*HeadlampIngressHostsPathsBackendServicePort)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampIngressTls",
		reflect.TypeOf((*HeadlampIngressTls)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampInitContainers",
		reflect.TypeOf((*HeadlampInitContainers)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampInitContainersEnv",
		reflect.TypeOf((*HeadlampInitContainersEnv)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"headlamp.HeadlampInitContainersImagePullPolicy",
		reflect.TypeOf((*HeadlampInitContainersImagePullPolicy)(nil)).Elem(),
		map[string]interface{}{
			"ALWAYS": HeadlampInitContainersImagePullPolicy_ALWAYS,
			"IF_NOT_PRESENT": HeadlampInitContainersImagePullPolicy_IF_NOT_PRESENT,
			"NEVER": HeadlampInitContainersImagePullPolicy_NEVER,
		},
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampInitContainersResources",
		reflect.TypeOf((*HeadlampInitContainersResources)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampInitContainersResourcesLimits",
		reflect.TypeOf((*HeadlampInitContainersResourcesLimits)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampInitContainersResourcesRequests",
		reflect.TypeOf((*HeadlampInitContainersResourcesRequests)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampInitContainersVolumeMounts",
		reflect.TypeOf((*HeadlampInitContainersVolumeMounts)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampPersistentVolumeClaim",
		reflect.TypeOf((*HeadlampPersistentVolumeClaim)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampPersistentVolumeClaimSelector",
		reflect.TypeOf((*HeadlampPersistentVolumeClaimSelector)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampPersistentVolumeClaimSelectorMatchExpressions",
		reflect.TypeOf((*HeadlampPersistentVolumeClaimSelectorMatchExpressions)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampPodDisruptionBudget",
		reflect.TypeOf((*HeadlampPodDisruptionBudget)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampProbes",
		reflect.TypeOf((*HeadlampProbes)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampProbesLivenessProbe",
		reflect.TypeOf((*HeadlampProbesLivenessProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampProbesReadinessProbe",
		reflect.TypeOf((*HeadlampProbesReadinessProbe)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"headlamp.HeadlampProbesScheme",
		reflect.TypeOf((*HeadlampProbesScheme)(nil)).Elem(),
		map[string]interface{}{
			"HTTP": HeadlampProbesScheme_HTTP,
			"HTTPS": HeadlampProbesScheme_HTTPS,
		},
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampProps",
		reflect.TypeOf((*HeadlampProps)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampService",
		reflect.TypeOf((*HeadlampService)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampServiceAccount",
		reflect.TypeOf((*HeadlampServiceAccount)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampServiceExtraServicePorts",
		reflect.TypeOf((*HeadlampServiceExtraServicePorts)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"headlamp.HeadlampServiceExtraServicePortsProtocol",
		reflect.TypeOf((*HeadlampServiceExtraServicePortsProtocol)(nil)).Elem(),
		map[string]interface{}{
			"TCP": HeadlampServiceExtraServicePortsProtocol_TCP,
			"UDP": HeadlampServiceExtraServicePortsProtocol_UDP,
			"SCTP": HeadlampServiceExtraServicePortsProtocol_SCTP,
		},
	)
	_jsii_.RegisterEnum(
		"headlamp.HeadlampServiceType",
		reflect.TypeOf((*HeadlampServiceType)(nil)).Elem(),
		map[string]interface{}{
			"CLUSTER_IP": HeadlampServiceType_CLUSTER_IP,
			"NODE_PORT": HeadlampServiceType_NODE_PORT,
			"LOAD_BALANCER": HeadlampServiceType_LOAD_BALANCER,
			"EXTERNAL_NAME": HeadlampServiceType_EXTERNAL_NAME,
		},
	)
	_jsii_.RegisterStruct(
		"headlamp.HeadlampValues",
		reflect.TypeOf((*HeadlampValues)(nil)).Elem(),
	)
}
