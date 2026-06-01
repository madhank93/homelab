package rancher


type RancherGatewayGatewayClassAdditionalListenersTls struct {
	CertificateRefs *[]*RancherGatewayGatewayClassAdditionalListenersTlsCertificateRefs `field:"optional" json:"certificateRefs" yaml:"certificateRefs"`
	Mode RancherGatewayGatewayClassAdditionalListenersTlsMode `field:"optional" json:"mode" yaml:"mode"`
}

