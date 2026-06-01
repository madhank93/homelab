package openbao


type OpenbaoServerIngress struct {
	ActiveService *bool `field:"optional" json:"activeService" yaml:"activeService"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	ExtraPaths *[]interface{} `field:"optional" json:"extraPaths" yaml:"extraPaths"`
	Hosts *[]*OpenbaoServerIngressHosts `field:"optional" json:"hosts" yaml:"hosts"`
	IngressClassName *string `field:"optional" json:"ingressClassName" yaml:"ingressClassName"`
	Labels interface{} `field:"optional" json:"labels" yaml:"labels"`
	PathType *string `field:"optional" json:"pathType" yaml:"pathType"`
	Tls *[]interface{} `field:"optional" json:"tls" yaml:"tls"`
}

