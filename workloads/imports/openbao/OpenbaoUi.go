package openbao


type OpenbaoUi struct {
	ActiveOpenbaoPodOnly *bool `field:"optional" json:"activeOpenbaoPodOnly" yaml:"activeOpenbaoPodOnly"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	Enabled interface{} `field:"optional" json:"enabled" yaml:"enabled"`
	ExternalPort *float64 `field:"optional" json:"externalPort" yaml:"externalPort"`
	ExternalTrafficPolicy *string `field:"optional" json:"externalTrafficPolicy" yaml:"externalTrafficPolicy"`
	PublishNotReadyAddresses *bool `field:"optional" json:"publishNotReadyAddresses" yaml:"publishNotReadyAddresses"`
	ServiceIpFamilies *[]interface{} `field:"optional" json:"serviceIpFamilies" yaml:"serviceIpFamilies"`
	ServiceIpFamilyPolicy *string `field:"optional" json:"serviceIpFamilyPolicy" yaml:"serviceIpFamilyPolicy"`
	ServiceNodePort *float64 `field:"optional" json:"serviceNodePort" yaml:"serviceNodePort"`
	ServiceType *string `field:"optional" json:"serviceType" yaml:"serviceType"`
	TargetPort *float64 `field:"optional" json:"targetPort" yaml:"targetPort"`
}

