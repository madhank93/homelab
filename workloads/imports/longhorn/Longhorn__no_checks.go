//go:build no_runtime_type_checking

package longhorn

// Building without runtime type checking enabled, so all the below just return nil

func validateLonghorn_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Longhorn) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewLonghornParameters(scope constructs.Construct, id *string, props *LonghornProps) error {
	return nil
}

