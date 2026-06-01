package openbao


type OpenbaoInjectorWebhook struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	FailurePolicy *string `field:"optional" json:"failurePolicy" yaml:"failurePolicy"`
	MatchPolicy *string `field:"optional" json:"matchPolicy" yaml:"matchPolicy"`
	NamespaceSelector interface{} `field:"optional" json:"namespaceSelector" yaml:"namespaceSelector"`
	ObjectSelector interface{} `field:"optional" json:"objectSelector" yaml:"objectSelector"`
	TimeoutSeconds *float64 `field:"optional" json:"timeoutSeconds" yaml:"timeoutSeconds"`
}

