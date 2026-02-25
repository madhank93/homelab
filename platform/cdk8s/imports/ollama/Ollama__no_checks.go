//go:build no_runtime_type_checking

package ollama

// Building without runtime type checking enabled, so all the below just return nil

func validateOllama_IsConstructParameters(x interface{}) error {
	return nil
}

func (j *jsiiProxy_Ollama) validateSetHelmParameters(val cdk8s.Helm) error {
	return nil
}

func validateNewOllamaParameters(scope constructs.Construct, id *string, props *OllamaProps) error {
	return nil
}

