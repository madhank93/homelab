package n8n


// This block is for setting up the ingress for more information can be found here: https://kubernetes.io/docs/concepts/services-networking/ingress/.
type N8NIngress struct {
	Annotations interface{} `field:"required" json:"annotations" yaml:"annotations"`
	ClassName *string `field:"required" json:"className" yaml:"className"`
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// kubernetes.io/ingress.class: nginx kubernetes.io/tls-acme: "true".
	Hosts *[]interface{} `field:"required" json:"hosts" yaml:"hosts"`
	Tls *[]interface{} `field:"required" json:"tls" yaml:"tls"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

