//go:build no_runtime_type_checking

package headlamp

// Building without runtime type checking enabled, so all the below just return nil

func validateHeadlamp_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Headlamp) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewHeadlampParameters(scope constructs.Construct, id *string, props *HeadlampProps) error {
	return nil
}

