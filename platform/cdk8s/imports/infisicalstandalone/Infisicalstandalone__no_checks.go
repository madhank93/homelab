//go:build no_runtime_type_checking

package infisicalstandalone

// Building without runtime type checking enabled, so all the below just return nil

func validateInfisicalstandalone_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Infisicalstandalone) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewInfisicalstandaloneParameters(scope constructs.Construct, id *string, props *InfisicalstandaloneProps) error {
	return nil
}

