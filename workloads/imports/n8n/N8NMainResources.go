package n8n


// This block is for setting up the resource management for the pod more information can be found here: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
type N8NMainResources struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Limits *N8NMainResourcesLimits `field:"optional" json:"limits" yaml:"limits"`
	Requests *N8NMainResourcesRequests `field:"optional" json:"requests" yaml:"requests"`
}

