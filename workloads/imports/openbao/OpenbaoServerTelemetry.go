package openbao


type OpenbaoServerTelemetry struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	PrometheusRules *OpenbaoServerTelemetryPrometheusRules `field:"optional" json:"prometheusRules" yaml:"prometheusRules"`
}

