package fleet

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/madhank93/homelab/workloads/imports/fleet/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"github.com/madhank93/homelab/workloads/imports/fleet/internal"
)

type Fleet interface {
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

// The jsii proxy struct for Fleet
type jsiiProxy_Fleet struct {
	internal.Type__constructsConstruct
}

func (j *jsiiProxy_Fleet) Helm() cdk8s.Helm {
	var returns cdk8s.Helm
	_jsii_.Get(
		j,
		"helm",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Fleet) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}


func NewFleet(scope constructs.Construct, id *string, props *FleetProps) Fleet {
	_init_.Initialize()

	if err := validateNewFleetParameters(scope, id, props); err != nil {
		panic(err)
	}
	j := jsiiProxy_Fleet{}

	_jsii_.Create(
		"fleet.Fleet",
		[]interface{}{scope, id, props},
		&j,
	)

	return &j
}

func NewFleet_Override(f Fleet, scope constructs.Construct, id *string, props *FleetProps) {
	_init_.Initialize()

	_jsii_.Create(
		"fleet.Fleet",
		[]interface{}{scope, id, props},
		f,
	)
}

func (j *jsiiProxy_Fleet)SetHelm(val cdk8s.Helm) {
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
func Fleet_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateFleet_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"fleet.Fleet",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func (f *jsiiProxy_Fleet) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		f,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}


