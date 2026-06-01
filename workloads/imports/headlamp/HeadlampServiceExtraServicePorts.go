package headlamp


type HeadlampServiceExtraServicePorts struct {
	// Port name (must be unique within the Service).
	Name *string `field:"required" json:"name" yaml:"name"`
	// Service port number.
	Port *float64 `field:"required" json:"port" yaml:"port"`
	// Node port (only honored when service.type is NodePort or LoadBalancer).
	NodePort *float64 `field:"optional" json:"nodePort" yaml:"nodePort"`
	// Protocol (TCP/UDP/SCTP).
	//
	// Defaults to TCP.
	// Default: TCP.
	//
	Protocol HeadlampServiceExtraServicePortsProtocol `field:"optional" json:"protocol" yaml:"protocol"`
	// Pod-side target port (number or named port).
	//
	// Defaults to `port` when omitted.
	// Default: port` when omitted.
	//
	TargetPort interface{} `field:"optional" json:"targetPort" yaml:"targetPort"`
}

