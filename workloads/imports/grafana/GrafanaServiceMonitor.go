package grafana


type GrafanaServiceMonitor struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	BasicAuth interface{} `field:"optional" json:"basicAuth" yaml:"basicAuth"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	MetricRelabelings *[]interface{} `field:"optional" json:"metricRelabelings" yaml:"metricRelabelings"`
	Path *string `field:"optional" json:"path" yaml:"path"`
	Relabelings *[]interface{} `field:"optional" json:"relabelings" yaml:"relabelings"`
	Scheme *string `field:"optional" json:"scheme" yaml:"scheme"`
	TargetLabels *[]interface{} `field:"optional" json:"targetLabels" yaml:"targetLabels"`
	TlsConfig interface{} `field:"optional" json:"tlsConfig" yaml:"tlsConfig"`
}

