package headlamp


type HeadlampIngress struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Annotations for Ingress resource.
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	// Enable ingress controller resource.
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Hosts *[]*HeadlampIngressHosts `field:"optional" json:"hosts" yaml:"hosts"`
	// Ingress class name.
	IngressClassName *string `field:"optional" json:"ingressClassName" yaml:"ingressClassName"`
	Tls *[]*HeadlampIngressTls `field:"optional" json:"tls" yaml:"tls"`
}

