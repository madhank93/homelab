package openbao


type OpenbaoInjectorAgentDefaults struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	CpuLimit *string `field:"optional" json:"cpuLimit" yaml:"cpuLimit"`
	CpuRequest *string `field:"optional" json:"cpuRequest" yaml:"cpuRequest"`
	EphemeralLimit *string `field:"optional" json:"ephemeralLimit" yaml:"ephemeralLimit"`
	EphemeralRequest *string `field:"optional" json:"ephemeralRequest" yaml:"ephemeralRequest"`
	MemLimit *string `field:"optional" json:"memLimit" yaml:"memLimit"`
	MemRequest *string `field:"optional" json:"memRequest" yaml:"memRequest"`
	Template *string `field:"optional" json:"template" yaml:"template"`
	TemplateConfig *OpenbaoInjectorAgentDefaultsTemplateConfig `field:"optional" json:"templateConfig" yaml:"templateConfig"`
}

