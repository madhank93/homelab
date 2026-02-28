package n8n


// DEPRECATED: Use main, worker, and webhook blocks resources fields instead.
//
// This field will be removed in a future release.
type N8NResources struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Limits *N8NResourcesLimits `field:"optional" json:"limits" yaml:"limits"`
	Requests *N8NResourcesRequests `field:"optional" json:"requests" yaml:"requests"`
}

