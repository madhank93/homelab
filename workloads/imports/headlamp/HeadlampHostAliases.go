package headlamp


type HeadlampHostAliases struct {
	// Hostnames that resolve to the given IP.
	Hostnames *[]*string `field:"required" json:"hostnames" yaml:"hostnames"`
	// IP address the hostnames resolve to.
	Ip *string `field:"required" json:"ip" yaml:"ip"`
}

