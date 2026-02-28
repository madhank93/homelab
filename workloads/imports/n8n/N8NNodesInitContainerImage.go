package n8n


// Image for the init container to install npm packages.
type N8NNodesInitContainerImage struct {
	// Pull policy for the init container to install npm packages.
	PullPolicy *string `field:"required" json:"pullPolicy" yaml:"pullPolicy"`
	// Repository for the init container to install npm packages.
	Repository *string `field:"required" json:"repository" yaml:"repository"`
	// Tag for the init container to install npm packages.
	Tag *string `field:"required" json:"tag" yaml:"tag"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

