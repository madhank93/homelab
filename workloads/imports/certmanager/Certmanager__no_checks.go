//go:build no_runtime_type_checking

package certmanager

// Building without runtime type checking enabled, so all the below just return nil

func validateCertmanager_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Certmanager) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewCertmanagerParameters(scope constructs.Construct, id *string, props *CertmanagerProps) error {
	return nil
}

