package n8n


type N8NTaskRunnersExternalResources struct {
	Limits *N8NTaskRunnersExternalResourcesLimits `field:"required" json:"limits" yaml:"limits"`
	Requests *N8NTaskRunnersExternalResourcesRequests `field:"required" json:"requests" yaml:"requests"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

