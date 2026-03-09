package openbao


type OpenbaoInjectorAgentDefaultsTemplateConfig struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	ExitOnRetryFailure *bool `field:"optional" json:"exitOnRetryFailure" yaml:"exitOnRetryFailure"`
	StaticSecretRenderInterval *string `field:"optional" json:"staticSecretRenderInterval" yaml:"staticSecretRenderInterval"`
}

