package headlamp


// Probe configuration for liveness and readiness checks.
type HeadlampProbes struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Liveness probe settings.
	LivenessProbe *HeadlampProbesLivenessProbe `field:"optional" json:"livenessProbe" yaml:"livenessProbe"`
	// Readiness probe settings.
	ReadinessProbe *HeadlampProbesReadinessProbe `field:"optional" json:"readinessProbe" yaml:"readinessProbe"`
	// Scheme for probes (HTTP or HTTPS).
	//
	// Set to HTTPS when TLS is enabled at the backend server.
	Scheme HeadlampProbesScheme `field:"optional" json:"scheme" yaml:"scheme"`
}

