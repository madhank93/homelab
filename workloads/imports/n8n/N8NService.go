package n8n


// This is for setting up a service more information can be found here: https://kubernetes.io/docs/concepts/services-networking/service/.
type N8NService struct {
	// Additional service annotations.
	Annotations interface{} `field:"required" json:"annotations" yaml:"annotations"`
	// Whether to enable the service.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Additional service labels.
	Labels interface{} `field:"required" json:"labels" yaml:"labels"`
	// Default Service name.
	Name *string `field:"required" json:"name" yaml:"name"`
	// This sets the ports more information can be found here: https://kubernetes.io/docs/concepts/services-networking/service/#field-spec-ports.
	Port *float64 `field:"required" json:"port" yaml:"port"`
	// This sets the service type more information can be found here: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types.
	Type *string `field:"required" json:"type" yaml:"type"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

