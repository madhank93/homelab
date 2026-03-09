// secrets-store-csi-driver
package secretsstorecsidriver

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterClass(
		"secrets-store-csi-driver.Secretsstorecsidriver",
		reflect.TypeOf((*Secretsstorecsidriver)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "helm", GoGetter: "Helm"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_Secretsstorecsidriver{}
			_jsii_.InitJsiiProxy(&j.Type__constructsConstruct)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"secrets-store-csi-driver.SecretsstorecsidriverProps",
		reflect.TypeOf((*SecretsstorecsidriverProps)(nil)).Elem(),
	)
}
