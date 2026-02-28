//go:build no_runtime_type_checking

package gpuoperator

// Building without runtime type checking enabled, so all the below just return nil

func validateGpuoperator_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Gpuoperator) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewGpuoperatorParameters(scope constructs.Construct, id *string, props *GpuoperatorProps) error {
	return nil
}

