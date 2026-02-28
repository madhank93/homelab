package n8n


// This block is for setting up the resource management for the pod more information can be found here: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
type N8NWorkerResources struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Limits *N8NWorkerResourcesLimits `field:"optional" json:"limits" yaml:"limits"`
	Requests *N8NWorkerResourcesRequests `field:"optional" json:"requests" yaml:"requests"`
}

