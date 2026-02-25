package n8n


type N8NTaskRunnersBroker struct {
	// This sets the address for the broker of the external task runner.
	Address *string `field:"required" json:"address" yaml:"address"`
	// This sets the port for the broker of the external task runner.
	Port *float64 `field:"required" json:"port" yaml:"port"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

