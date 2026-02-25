package infisicalstandalone

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/madhank93/homelab/cdk8s/imports/infisicalstandalone/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/cdk8s/imports/infisicalstandalone/internal"
)

type Infisicalstandalone interface {
	constructs.Construct
	Helm() cdk8s.Helm
	SetHelm(val cdk8s.Helm)
	// The tree node.
	Node() constructs.Node
	// Returns a string representation of this construct.
	ToString() *string
	// Applies one or more mixins to this construct.
	//
	// Mixins are applied in order. The list of constructs is captured at the
	// start of the call, so constructs added by a mixin will not be visited.
	// Use multiple `with()` calls if subsequent mixins should apply to added
	// constructs.
	//
	// Returns: This construct for chaining.
}

// The jsii proxy struct for Infisicalstandalone
type jsiiProxy_Infisicalstandalone struct {
	internal.Type__constructsConstruct
}

func (j *jsiiProxy_Infisicalstandalone) Helm() cdk8s.Helm {
	var returns cdk8s.Helm
	_jsii_.Get(
		j,
		"helm",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Infisicalstandalone) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}


func NewInfisicalstandalone(scope constructs.Construct, id *string, props *InfisicalstandaloneProps) Infisicalstandalone {
	_init_.Initialize()

	if err := validateNewInfisicalstandaloneParameters(scope, id, props); err != nil {
		panic(err)
	}
	j := jsiiProxy_Infisicalstandalone{}

	_jsii_.Create(
		"infisical-standalone.Infisicalstandalone",
		[]interface{}{scope, id, props},
		&j,
	)

	return &j
}

func NewInfisicalstandalone_Override(i Infisicalstandalone, scope constructs.Construct, id *string, props *InfisicalstandaloneProps) {
	_init_.Initialize()

	_jsii_.Create(
		"infisical-standalone.Infisicalstandalone",
		[]interface{}{scope, id, props},
		i,
	)
}

func (j *jsiiProxy_Infisicalstandalone)SetHelm(val cdk8s.Helm) {
	if err := j.validateSetHelmParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"helm",
		val,
	)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct`
// instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on
// disk are seen as independent, completely different libraries. As a
// consequence, the class `Construct` in each copy of the `constructs` library
// is seen as a different class, and an instance of one class will not test as
// `instanceof` the other class. `npm install` will not create installations
// like this, but users may manually symlink construct libraries together or
// use a monorepo tool: in those cases, multiple copies of the `constructs`
// library can be accidentally installed, and `instanceof` will behave
// unpredictably. It is safest to avoid using `instanceof`, and using
// this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Infisicalstandalone_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateInfisicalstandalone_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"infisical-standalone.Infisicalstandalone",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func (i *jsiiProxy_Infisicalstandalone) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		i,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}


