package n8n


// This block is for setting up the resource management for the init container more information can be found here: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
type N8NNodesInitContainerResourcesLimits struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// CPU for the init container to install npm packages.
	Cpu *string `field:"optional" json:"cpu" yaml:"cpu"`
	// Memory for the init container to install npm packages.
	Memory *string `field:"optional" json:"memory" yaml:"memory"`
}

