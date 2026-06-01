//go:build no_runtime_type_checking

package secretsstorecsixk8sio

// Building without runtime type checking enabled, so all the below just return nil

func validateSecretProviderClass_IsApiObjectParameters(o interface{}) error {
	return nil
}

func validateSecretProviderClass_IsConstructParameters(x interface{}) error {
	return nil
}

func validateSecretProviderClass_ManifestParameters(props *SecretProviderClassProps) error {
	return nil
}

func validateSecretProviderClass_OfParameters(c constructs.IConstruct) error {
	return nil
}

func validateNewSecretProviderClassParameters(scope constructs.Construct, id *string, props *SecretProviderClassProps) error {
	return nil
}

