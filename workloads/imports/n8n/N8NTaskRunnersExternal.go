package n8n


type N8NTaskRunnersExternal struct {
	// This sets the auto shutdown timeout for the external task runner in seconds.
	AutoShutdownTimeout *float64 `field:"required" json:"autoShutdownTimeout" yaml:"autoShutdownTimeout"`
	// This sets the node options for the external task runner.
	NodeOptions *[]*string `field:"required" json:"nodeOptions" yaml:"nodeOptions"`
	// This sets the ports for the external task runner more information can be found here: https://kubernetes.io/docs/concepts/services-networking/service/#field-spec-ports.
	Port *float64 `field:"required" json:"port" yaml:"port"`
	Resources *N8NTaskRunnersExternalResources `field:"required" json:"resources" yaml:"resources"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// This sets the auth token for the n8n main node of the external task runner.
	MainNodeAuthToken *string `field:"optional" json:"mainNodeAuthToken" yaml:"mainNodeAuthToken"`
	// This sets the auth token for the n8n worker node of the external task runner.
	WorkerNodeAuthToken *string `field:"optional" json:"workerNodeAuthToken" yaml:"workerNodeAuthToken"`
}

