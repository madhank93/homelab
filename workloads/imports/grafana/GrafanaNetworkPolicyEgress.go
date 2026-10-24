package grafana


type GrafanaNetworkPolicyEgress struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// When enabled, DNS resolution will be blocked for all pods in the grafana namespace.
	BlockDnsResolution *bool `field:"optional" json:"blockDnsResolution" yaml:"blockDnsResolution"`
	// When enabled, an egress network policy will be created allowing grafana to connect to external data sources from kubernetes cluster.
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	// Add individual ports to be allowed by the egress.
	Ports *[]interface{} `field:"optional" json:"ports" yaml:"ports"`
	// Allow egress traffic to specific destinations.
	To *[]interface{} `field:"optional" json:"to" yaml:"to"`
}

