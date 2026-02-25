//go:build no_runtime_type_checking

package kyverno

// Building without runtime type checking enabled, so all the below just return nil

func validateKyverno_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Kyverno) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewKyvernoParameters(scope constructs.Construct, id *string, props *KyvernoProps) error {
	return nil
}

