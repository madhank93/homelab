//go:build no_runtime_type_checking

package grafana

// Building without runtime type checking enabled, so all the below just return nil

func validateGrafana_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Grafana) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewGrafanaParameters(scope constructs.Construct, id *string, props *GrafanaProps) error {
	return nil
}

