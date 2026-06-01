package headlamp


// HTTPRoute configuration for Gateway API.
type HeadlampHttpRoute struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Annotations for HTTPRoute resource.
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	// Enable HTTPRoute resource for Gateway API.
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	// Hostnames for the HTTPRoute.
	Hostnames *[]*string `field:"optional" json:"hostnames" yaml:"hostnames"`
	// Additional labels for HTTPRoute resource.
	Labels interface{} `field:"optional" json:"labels" yaml:"labels"`
	// Parent references (REQUIRED when enabled - HTTPRoute will not work without this).
	ParentRefs *[]*HeadlampHttpRouteParentRefs `field:"optional" json:"parentRefs" yaml:"parentRefs"`
	// Custom routing rules (optional, defaults to path prefix /).
	Rules *[]interface{} `field:"optional" json:"rules" yaml:"rules"`
}

