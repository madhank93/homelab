package n8n


// Image for the init container to install npm packages.
type N8NNodesInitContainer struct {
	// Image for the init container to install npm packages.
	Image *N8NNodesInitContainerImage `field:"required" json:"image" yaml:"image"`
	// This block is for setting up the resource management for the init container more information can be found here: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
	Resources *N8NNodesInitContainerResources `field:"required" json:"resources" yaml:"resources"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

