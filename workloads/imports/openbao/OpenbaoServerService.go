package openbao


type OpenbaoServerService struct {
	Active *OpenbaoServerServiceActive `field:"optional" json:"active" yaml:"active"`
	ActiveNodePort *float64 `field:"optional" json:"activeNodePort" yaml:"activeNodePort"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	ExternalTrafficPolicy *string `field:"optional" json:"externalTrafficPolicy" yaml:"externalTrafficPolicy"`
	InstanceSelector *OpenbaoServerServiceInstanceSelector `field:"optional" json:"instanceSelector" yaml:"instanceSelector"`
	IpFamilies *[]interface{} `field:"optional" json:"ipFamilies" yaml:"ipFamilies"`
	IpFamilyPolicy *string `field:"optional" json:"ipFamilyPolicy" yaml:"ipFamilyPolicy"`
	NodePort *float64 `field:"optional" json:"nodePort" yaml:"nodePort"`
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	PublishNotReadyAddresses *bool `field:"optional" json:"publishNotReadyAddresses" yaml:"publishNotReadyAddresses"`
	Standby *OpenbaoServerServiceStandby `field:"optional" json:"standby" yaml:"standby"`
	StandbyNodePort *float64 `field:"optional" json:"standbyNodePort" yaml:"standbyNodePort"`
	TargetPort *float64 `field:"optional" json:"targetPort" yaml:"targetPort"`
}

