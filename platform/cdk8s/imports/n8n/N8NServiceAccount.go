package n8n


// This section builds out the service account more information can be found here: https://kubernetes.io/docs/concepts/security/service-accounts/.
type N8NServiceAccount struct {
	// Annotations to add to the service account.
	Annotations interface{} `field:"required" json:"annotations" yaml:"annotations"`
	// Automatically mount a ServiceAccount's API credentials?
	Automount *bool `field:"required" json:"automount" yaml:"automount"`
	// Specifies whether a service account should be created.
	Create *bool `field:"required" json:"create" yaml:"create"`
	// The name of the service account to use.
	//
	// If not set and create is true, a name is generated using the fullname template.
	Name *string `field:"required" json:"name" yaml:"name"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

