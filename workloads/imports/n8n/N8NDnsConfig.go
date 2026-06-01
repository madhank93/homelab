package n8n


// For more information checkout: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-dns-config.
type N8NDnsConfig struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// List of DNS nameserver IP addresses (IPv4 or IPv6).
	Nameservers *[]*string `field:"optional" json:"nameservers" yaml:"nameservers"`
	// List of DNS resolver options.
	Options *[]*N8NDnsConfigOptions `field:"optional" json:"options" yaml:"options"`
	// List of DNS search domains for hostname lookup.
	Searches *[]*string `field:"optional" json:"searches" yaml:"searches"`
}

