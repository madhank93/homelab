package grafana


type GrafanaIngress struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations *map[string]*string `field:"optional" json:"annotations" yaml:"annotations"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	ExtraPaths *[]interface{} `field:"optional" json:"extraPaths" yaml:"extraPaths"`
	Hosts *[]*string `field:"optional" json:"hosts" yaml:"hosts"`
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	Path *string `field:"optional" json:"path" yaml:"path"`
	PathType *string `field:"optional" json:"pathType" yaml:"pathType"`
	Tls *[]interface{} `field:"optional" json:"tls" yaml:"tls"`
}

