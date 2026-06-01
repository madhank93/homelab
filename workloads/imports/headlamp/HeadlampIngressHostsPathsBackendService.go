package headlamp


type HeadlampIngressHostsPathsBackendService struct {
	// Service name (supports tpl).
	//
	// Defaults to the Headlamp Service when omitted.
	// Default: the Headlamp Service when omitted.
	//
	Name *string `field:"optional" json:"name" yaml:"name"`
	Port *HeadlampIngressHostsPathsBackendServicePort `field:"optional" json:"port" yaml:"port"`
}

