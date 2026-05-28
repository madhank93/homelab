// victoria-metrics-k8s-stack
package victoriametricsk8sstack

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterClass(
		"victoria-metrics-k8s-stack.Victoriametricsk8sstack",
		reflect.TypeOf((*Victoriametricsk8sstack)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "helm", GoGetter: "Helm"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_Victoriametricsk8sstack{}
			_jsii_.InitJsiiProxy(&j.Type__constructsConstruct)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"victoria-metrics-k8s-stack.Victoriametricsk8sstackProps",
		reflect.TypeOf((*Victoriametricsk8sstackProps)(nil)).Elem(),
	)
}
