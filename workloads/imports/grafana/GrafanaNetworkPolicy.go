package grafana


type GrafanaNetworkPolicy struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Don't require client label for connections.
	//
	// The Policy model to apply. When set to false, only pods with the correct client label will have network access to grafana port defined. When true, grafana will accept connections from any source (with the correct destination port).
	AllowExternal *bool `field:"optional" json:"allowExternal" yaml:"allowExternal"`
	Egress *GrafanaNetworkPolicyEgress `field:"optional" json:"egress" yaml:"egress"`
	// Enable creation of NetworkPolicy resources.
	//
	// Only Ingress traffic is filtered for now.
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	// List of CIDR blocks allowed as ingress sources.
	//
	// Each entry must be a valid CIDR notation string (e.g. 10.0.0.0/8). When defined, the specified CIDR ranges are added to the ingress `from` rules using `ipBlock` entries and complement the other configured ingress sources. </br>  Example:  ``` explicitIpBlocks: - 35.191.0.0/16 - 130.211.0.0/22 ```
	ExplicitIpBlocks *[]interface{} `field:"optional" json:"explicitIpBlocks" yaml:"explicitIpBlocks"`
	// A Kubernetes LabelSelector to explicitly select namespaces from which traffic could be allowed If explicitNamespacesSelector is missing or set to {}, only client Pods that are in the networkPolicy's namespace and that match other criteria, the ones that have the good label, can reach the grafana.
	//
	// But sometimes, we want the grafana to be accessible to clients from other namespaces, in this case, we can use this LabelSelector to select these namespaces, note that the networkPolicy's namespace should also be explicitly added. </br>  Example:  ``` explicitNamespacesSelector: matchLabels: role: frontend matchExpressions: - {key: role, operator: In, values: [frontend]} ```
	ExplicitNamespacesSelector interface{} `field:"optional" json:"explicitNamespacesSelector" yaml:"explicitNamespacesSelector"`
	// When true enables the creation of an ingress network policy.
	Ingress *bool `field:"optional" json:"ingress" yaml:"ingress"`
}

