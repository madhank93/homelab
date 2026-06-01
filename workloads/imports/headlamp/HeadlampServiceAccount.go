package headlamp


type HeadlampServiceAccount struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Annotations to add to the service account.
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	// Specifies whether a service account should be created.
	Create *bool `field:"optional" json:"create" yaml:"create"`
	// The name of the service account to use.
	Name *string `field:"optional" json:"name" yaml:"name"`
}

