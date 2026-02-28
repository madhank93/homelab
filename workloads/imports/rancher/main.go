// rancher
package rancher

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterClass(
		"rancher.Rancher",
		reflect.TypeOf((*Rancher)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "helm", GoGetter: "Helm"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "with", GoMethod: "With"},
		},
		func() interface{} {
			j := jsiiProxy_Rancher{}
			_jsii_.InitJsiiProxy(&j.Type__constructsConstruct)
			return &j
		},
	)
	_jsii_.RegisterEnum(
		"rancher.RancherAgentTlsMode",
		reflect.TypeOf((*RancherAgentTlsMode)(nil)).Elem(),
		map[string]interface{}{
			"STRICT": RancherAgentTlsMode_STRICT,
			"SYSTEM_HYPHEN_STORE": RancherAgentTlsMode_SYSTEM_HYPHEN_STORE,
		},
	)
	_jsii_.RegisterStruct(
		"rancher.RancherAuditLog",
		reflect.TypeOf((*RancherAuditLog)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"rancher.RancherAuditLogDestination",
		reflect.TypeOf((*RancherAuditLogDestination)(nil)).Elem(),
		map[string]interface{}{
			"SIDECAR": RancherAuditLogDestination_SIDECAR,
			"HOST_PATH": RancherAuditLogDestination_HOST_PATH,
		},
	)
	_jsii_.RegisterEnum(
		"rancher.RancherAuditLogLevel",
		reflect.TypeOf((*RancherAuditLogLevel)(nil)).Elem(),
		map[string]interface{}{
			"VALUE_0": RancherAuditLogLevel_VALUE_0,
			"VALUE_1": RancherAuditLogLevel_VALUE_1,
			"VALUE_2": RancherAuditLogLevel_VALUE_2,
			"VALUE_3": RancherAuditLogLevel_VALUE_3,
		},
	)
	_jsii_.RegisterStruct(
		"rancher.RancherIngress",
		reflect.TypeOf((*RancherIngress)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"rancher.RancherIngressServicePort",
		reflect.TypeOf((*RancherIngressServicePort)(nil)).Elem(),
		map[string]interface{}{
			"VALUE_443": RancherIngressServicePort_VALUE_443,
			"VALUE_80": RancherIngressServicePort_VALUE_80,
		},
	)
	_jsii_.RegisterStruct(
		"rancher.RancherProps",
		reflect.TypeOf((*RancherProps)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"rancher.RancherService",
		reflect.TypeOf((*RancherService)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"rancher.RancherServiceType",
		reflect.TypeOf((*RancherServiceType)(nil)).Elem(),
		map[string]interface{}{
			"CLUSTER_IP": RancherServiceType_CLUSTER_IP,
			"LOAD_BALANCER": RancherServiceType_LOAD_BALANCER,
			"NODE_PORT": RancherServiceType_NODE_PORT,
		},
	)
	_jsii_.RegisterStruct(
		"rancher.RancherValues",
		reflect.TypeOf((*RancherValues)(nil)).Elem(),
	)
}
