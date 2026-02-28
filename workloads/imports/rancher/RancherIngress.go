package rancher


// The default rancher ingress configuration.
type RancherIngress struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	ServicePort RancherIngressServicePort `field:"optional" json:"servicePort" yaml:"servicePort"`
}

