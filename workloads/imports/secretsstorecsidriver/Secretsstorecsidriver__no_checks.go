//go:build no_runtime_type_checking

package secretsstorecsidriver

// Building without runtime type checking enabled, so all the below just return nil

func validateSecretsstorecsidriver_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Secretsstorecsidriver) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewSecretsstorecsidriverParameters(scope constructs.Construct, id *string, props *SecretsstorecsidriverProps) error {
	return nil
}

