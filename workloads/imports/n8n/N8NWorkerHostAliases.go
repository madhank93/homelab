package n8n


type N8NWorkerHostAliases struct {
	// List of hostnames to associate with the IP address.
	Hostnames *[]*string `field:"required" json:"hostnames" yaml:"hostnames"`
	// IP address for the host alias.
	Ip *string `field:"required" json:"ip" yaml:"ip"`
}

