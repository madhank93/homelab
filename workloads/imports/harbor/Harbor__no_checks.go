//go:build no_runtime_type_checking

package harbor

// Building without runtime type checking enabled, so all the below just return nil

func validateHarbor_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Harbor) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewHarborParameters(scope constructs.Construct, id *string, props *HarborProps) error {
	return nil
}

