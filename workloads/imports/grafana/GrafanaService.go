package grafana


type GrafanaService struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations *map[string]*string `field:"optional" json:"annotations" yaml:"annotations"`
	AppProtocol *string `field:"optional" json:"appProtocol" yaml:"appProtocol"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	IpFamilies *[]interface{} `field:"optional" json:"ipFamilies" yaml:"ipFamilies"`
	IpFamilyPolicy *string `field:"optional" json:"ipFamilyPolicy" yaml:"ipFamilyPolicy"`
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	LoadBalancerClass *string `field:"optional" json:"loadBalancerClass" yaml:"loadBalancerClass"`
	LoadBalancerIp *string `field:"optional" json:"loadBalancerIp" yaml:"loadBalancerIp"`
	LoadBalancerSourceRanges *[]interface{} `field:"optional" json:"loadBalancerSourceRanges" yaml:"loadBalancerSourceRanges"`
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	PortName *string `field:"optional" json:"portName" yaml:"portName"`
	SessionAffinity *string `field:"optional" json:"sessionAffinity" yaml:"sessionAffinity"`
	TargetPort *float64 `field:"optional" json:"targetPort" yaml:"targetPort"`
	TrafficDistribution *string `field:"optional" json:"trafficDistribution" yaml:"trafficDistribution"`
	Type *string `field:"optional" json:"type" yaml:"type"`
}

