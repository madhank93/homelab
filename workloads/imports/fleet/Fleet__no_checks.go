//go:build no_runtime_type_checking

package fleet

// Building without runtime type checking enabled, so all the below just return nil

func validateFleet_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Fleet) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewFleetParameters(scope constructs.Construct, id *string, props *FleetProps) error {
	return nil
}

