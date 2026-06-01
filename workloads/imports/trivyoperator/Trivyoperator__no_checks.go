//go:build no_runtime_type_checking

package trivyoperator

// Building without runtime type checking enabled, so all the below just return nil

func validateTrivyoperator_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Trivyoperator) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewTrivyoperatorParameters(scope constructs.Construct, id *string, props *TrivyoperatorProps) error {
	return nil
}

