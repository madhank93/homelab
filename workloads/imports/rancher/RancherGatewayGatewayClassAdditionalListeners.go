package rancher


type RancherGatewayGatewayClassAdditionalListeners struct {
	Name *string `field:"required" json:"name" yaml:"name"`
	Port *float64 `field:"required" json:"port" yaml:"port"`
	Protocol RancherGatewayGatewayClassAdditionalListenersProtocol `field:"required" json:"protocol" yaml:"protocol"`
	Hostname *string `field:"optional" json:"hostname" yaml:"hostname"`
	Tls *RancherGatewayGatewayClassAdditionalListenersTls `field:"optional" json:"tls" yaml:"tls"`
}

