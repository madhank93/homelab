//go:build no_runtime_type_checking

package metricsserver

// Building without runtime type checking enabled, so all the below just return nil

func validateMetricsserver_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Metricsserver) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewMetricsserverParameters(scope constructs.Construct, id *string, props *MetricsserverProps) error {
	return nil
}

