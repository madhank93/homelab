package n8n


// This block is for setting up the resource management for the init container more information can be found here: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
type N8NNodesInitContainerResources struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// This block is for setting up the resource management for the init container more information can be found here: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
	Limits *N8NNodesInitContainerResourcesLimits `field:"optional" json:"limits" yaml:"limits"`
	// This block is for setting up the resource management for the init container more information can be found here: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
	Requests *N8NNodesInitContainerResourcesRequests `field:"optional" json:"requests" yaml:"requests"`
}

