//go:build no_runtime_type_checking

package victoriametricscluster

// Building without runtime type checking enabled, so all the below just return nil

func validateVictoriametricscluster_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Victoriametricscluster) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewVictoriametricsclusterParameters(scope constructs.Construct, id *string, props *VictoriametricsclusterProps) error {
	return nil
}

