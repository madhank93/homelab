package grafana


type GrafanaAdmin struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	ExistingSecret *string `field:"optional" json:"existingSecret" yaml:"existingSecret"`
	PasswordKey *string `field:"optional" json:"passwordKey" yaml:"passwordKey"`
	UserKey *string `field:"optional" json:"userKey" yaml:"userKey"`
}

