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
		"headlamp.HeadlampEnv",
		reflect.TypeOf((*HeadlampEnv)(nil)).Elem(),
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
