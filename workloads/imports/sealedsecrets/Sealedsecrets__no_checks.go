//go:build no_runtime_type_checking

package sealedsecrets

// Building without runtime type checking enabled, so all the below just return nil

func validateSealedsecrets_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Sealedsecrets) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewSealedsecretsParameters(scope constructs.Construct, id *string, props *SealedsecretsProps) error {
	return nil
}

