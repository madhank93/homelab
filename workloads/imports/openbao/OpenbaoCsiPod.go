package openbao


type OpenbaoCsiPod struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Affinity interface{} `field:"optional" json:"affinity" yaml:"affinity"`
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	ExtraLabels interface{} `field:"optional" json:"extraLabels" yaml:"extraLabels"`
	NodeSelector interface{} `field:"optional" json:"nodeSelector" yaml:"nodeSelector"`
	Tolerations interface{} `field:"optional" json:"tolerations" yaml:"tolerations"`
}

