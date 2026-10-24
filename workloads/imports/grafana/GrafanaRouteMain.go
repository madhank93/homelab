package grafana


type GrafanaRouteMain struct {
	AdditionalRules *[]interface{} `field:"optional" json:"additionalRules" yaml:"additionalRules"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Annotations *map[string]*string `field:"optional" json:"annotations" yaml:"annotations"`
	// Set the route apiVersion, e.g. gateway.networking.k8s.io/v1 or gateway.networking.k8s.io/v1alpha2.
	ApiVersion *string `field:"optional" json:"apiVersion" yaml:"apiVersion"`
	// Enables or disables the route.
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Filters *[]interface{} `field:"optional" json:"filters" yaml:"filters"`
	Hostnames *[]interface{} `field:"optional" json:"hostnames" yaml:"hostnames"`
	HttpsRedirect *bool `field:"optional" json:"httpsRedirect" yaml:"httpsRedirect"`
	// Set the route kind Valid options are GRPCRoute, HTTPRoute, TCPRoute, TLSRoute, UDPRoute.
	Kind *string `field:"optional" json:"kind" yaml:"kind"`
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	Matches *[]*GrafanaRouteMainMatches `field:"optional" json:"matches" yaml:"matches"`
	ParentRefs *[]interface{} `field:"optional" json:"parentRefs" yaml:"parentRefs"`
	SessionPersistence interface{} `field:"optional" json:"sessionPersistence" yaml:"sessionPersistence"`
	Timeouts interface{} `field:"optional" json:"timeouts" yaml:"timeouts"`
}

