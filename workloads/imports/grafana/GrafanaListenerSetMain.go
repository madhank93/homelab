package grafana


type GrafanaListenerSetMain struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations *map[string]*string `field:"optional" json:"annotations" yaml:"annotations"`
	// Set the ListenerSet apiVersion, e.g. gateway.networking.k8s.io/v1.
	ApiVersion *string `field:"optional" json:"apiVersion" yaml:"apiVersion"`
	// Enables or disables the listener set.
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// List of listeners to attach to the parent Gateway.
	Listeners *[]interface{} `field:"optional" json:"listeners" yaml:"listeners"`
	// Reference to the parent Gateway this ListenerSet attaches to.
	ParentRef interface{} `field:"optional" json:"parentRef" yaml:"parentRef"`
}

