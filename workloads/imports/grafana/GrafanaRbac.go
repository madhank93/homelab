package grafana


type GrafanaRbac struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Create *bool `field:"optional" json:"create" yaml:"create"`
	ExtraClusterRoleRules *[]interface{} `field:"optional" json:"extraClusterRoleRules" yaml:"extraClusterRoleRules"`
	ExtraRoleRules *[]interface{} `field:"optional" json:"extraRoleRules" yaml:"extraRoleRules"`
	Namespaced *bool `field:"optional" json:"namespaced" yaml:"namespaced"`
	Namespaces *[]interface{} `field:"optional" json:"namespaces" yaml:"namespaces"`
	PspEnabled *bool `field:"optional" json:"pspEnabled" yaml:"pspEnabled"`
	PspUseAppArmor *bool `field:"optional" json:"pspUseAppArmor" yaml:"pspUseAppArmor"`
}

