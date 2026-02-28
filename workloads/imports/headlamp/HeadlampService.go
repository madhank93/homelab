package headlamp


type HeadlampService struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Kubernetes Service clusterIP.
	ClusterIp *string `field:"optional" json:"clusterIp" yaml:"clusterIp"`
	// Kubernetes Service loadBalancerIP.
	LoadBalancerIp *string `field:"optional" json:"loadBalancerIp" yaml:"loadBalancerIp"`
	// Kubernetes Service loadBalancerSourceRanges.
	LoadBalancerSourceRanges *[]*string `field:"optional" json:"loadBalancerSourceRanges" yaml:"loadBalancerSourceRanges"`
	// Kubernetes Service Nodeport.
	NodePort *float64 `field:"optional" json:"nodePort" yaml:"nodePort"`
	// Kubernetes Service port.
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	// Kubernetes Service type.
	Type HeadlampServiceType `field:"optional" json:"type" yaml:"type"`
}

