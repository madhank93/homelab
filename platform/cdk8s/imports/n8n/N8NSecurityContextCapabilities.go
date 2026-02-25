package n8n


type N8NSecurityContextCapabilities struct {
	Add *[]interface{} `field:"optional" json:"add" yaml:"add"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Drop *[]interface{} `field:"optional" json:"drop" yaml:"drop"`
}

