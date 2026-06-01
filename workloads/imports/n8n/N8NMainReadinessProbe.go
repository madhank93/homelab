package n8n


// This is to setup the readiness probe for the main pod more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/.
type N8NMainReadinessProbe struct {
	HttpGet *N8NMainReadinessProbeHttpGet `field:"required" json:"httpGet" yaml:"httpGet"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

