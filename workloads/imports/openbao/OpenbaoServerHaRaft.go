package openbao


type OpenbaoServerHaRaft struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Config interface{} `field:"optional" json:"config" yaml:"config"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	SetNodeId *bool `field:"optional" json:"setNodeId" yaml:"setNodeId"`
}

