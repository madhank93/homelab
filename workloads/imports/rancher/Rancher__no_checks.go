//go:build no_runtime_type_checking

package rancher

// Building without runtime type checking enabled, so all the below just return nil

func validateRancher_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Rancher) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewRancherParameters(scope constructs.Construct, id *string, props *RancherProps) error {
	return nil
}

