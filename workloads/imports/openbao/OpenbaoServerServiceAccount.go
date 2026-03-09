package openbao


type OpenbaoServerServiceAccount struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	Create *bool `field:"optional" json:"create" yaml:"create"`
	CreateSecret *bool `field:"optional" json:"createSecret" yaml:"createSecret"`
	ExtraLabels interface{} `field:"optional" json:"extraLabels" yaml:"extraLabels"`
	Name *string `field:"optional" json:"name" yaml:"name"`
	ServiceDiscovery *OpenbaoServerServiceAccountServiceDiscovery `field:"optional" json:"serviceDiscovery" yaml:"serviceDiscovery"`
}

