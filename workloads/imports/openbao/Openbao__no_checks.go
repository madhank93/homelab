//go:build no_runtime_type_checking

package openbao

// Building without runtime type checking enabled, so all the below just return nil

func validateOpenbao_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Openbao) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewOpenbaoParameters(scope constructs.Construct, id *string, props *OpenbaoProps) error {
	return nil
}

