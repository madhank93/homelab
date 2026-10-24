package grafana


type GrafanaServiceAccount struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	AutomountServiceAccountToken *bool `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	Create *bool `field:"optional" json:"create" yaml:"create"`
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	Name *string `field:"optional" json:"name" yaml:"name"`
	NameTest *string `field:"optional" json:"nameTest" yaml:"nameTest"`
}

