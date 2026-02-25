//go:build no_runtime_type_checking

package kubeprometheusstack

// Building without runtime type checking enabled, so all the below just return nil

func validateKubeprometheusstack_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Kubeprometheusstack) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewKubeprometheusstackParameters(scope constructs.Construct, id *string, props *KubeprometheusstackProps) error {
	return nil
}

