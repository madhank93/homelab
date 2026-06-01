package n8n


// DEPRECATED: Use main, worker, and webhook blocks livenessProbe field instead.
//
// This field will be removed in a future release.
type N8NLivenessProbe struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	HttpGet *N8NLivenessProbeHttpGet `field:"optional" json:"httpGet" yaml:"httpGet"`
}

