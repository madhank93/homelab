//go:build no_runtime_type_checking

package n8n

// Building without runtime type checking enabled, so all the below just return nil

func validateN8n_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_N8n) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewN8nParameters(scope constructs.Construct, id *string, props *N8nProps) error {
	return nil
}

