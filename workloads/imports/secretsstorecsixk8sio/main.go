// secrets-storecsix-k8sio
package secretsstorecsixk8sio

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterClass(
		"secrets-storecsix-k8sio.SecretProviderClass",
		reflect.TypeOf((*SecretProviderClass)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addDependency", GoMethod: "AddDependency"},
			_jsii_.MemberMethod{JsiiMethod: "addJsonPatch", GoMethod: "AddJsonPatch"},
			_jsii_.MemberProperty{JsiiProperty: "apiGroup", GoGetter: "ApiGroup"},
			_jsii_.MemberProperty{JsiiProperty: "apiVersion", GoGetter: "ApiVersion"},
			_jsii_.MemberProperty{JsiiProperty: "chart", GoGetter: "Chart"},
			_jsii_.MemberProperty{JsiiProperty: "kind", GoGetter: "Kind"},
			_jsii_.MemberProperty{JsiiProperty: "metadata", GoGetter: "Metadata"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "toJson", GoMethod: "ToJson"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_SecretProviderClass{}
			_jsii_.InitJsiiProxy(&j.Type__cdk8sApiObject)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"secrets-storecsix-k8sio.SecretProviderClassProps",
		reflect.TypeOf((*SecretProviderClassProps)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"secrets-storecsix-k8sio.SecretProviderClassSpec",
		reflect.TypeOf((*SecretProviderClassSpec)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"secrets-storecsix-k8sio.SecretProviderClassSpecSecretObjects",
		reflect.TypeOf((*SecretProviderClassSpecSecretObjects)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"secrets-storecsix-k8sio.SecretProviderClassSpecSecretObjectsData",
		reflect.TypeOf((*SecretProviderClassSpecSecretObjectsData)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"secrets-storecsix-k8sio.SecretProviderClassV1Alpha1",
		reflect.TypeOf((*SecretProviderClassV1Alpha1)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addDependency", GoMethod: "AddDependency"},
			_jsii_.MemberMethod{JsiiMethod: "addJsonPatch", GoMethod: "AddJsonPatch"},
			_jsii_.MemberProperty{JsiiProperty: "apiGroup", GoGetter: "ApiGroup"},
			_jsii_.MemberProperty{JsiiProperty: "apiVersion", GoGetter: "ApiVersion"},
			_jsii_.MemberProperty{JsiiProperty: "chart", GoGetter: "Chart"},
			_jsii_.MemberProperty{JsiiProperty: "kind", GoGetter: "Kind"},
			_jsii_.MemberProperty{JsiiProperty: "metadata", GoGetter: "Metadata"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "toJson", GoMethod: "ToJson"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_SecretProviderClassV1Alpha1{}
			_jsii_.InitJsiiProxy(&j.Type__cdk8sApiObject)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"secrets-storecsix-k8sio.SecretProviderClassV1Alpha1Props",
		reflect.TypeOf((*SecretProviderClassV1Alpha1Props)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"secrets-storecsix-k8sio.SecretProviderClassV1Alpha1Spec",
		reflect.TypeOf((*SecretProviderClassV1Alpha1Spec)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"secrets-storecsix-k8sio.SecretProviderClassV1Alpha1SpecSecretObjects",
		reflect.TypeOf((*SecretProviderClassV1Alpha1SpecSecretObjects)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"secrets-storecsix-k8sio.SecretProviderClassV1Alpha1SpecSecretObjectsData",
		reflect.TypeOf((*SecretProviderClassV1Alpha1SpecSecretObjectsData)(nil)).Elem(),
	)
}
