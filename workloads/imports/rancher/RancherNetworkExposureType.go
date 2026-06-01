package rancher


type RancherNetworkExposureType string

const (
	// ingress.
	RancherNetworkExposureType_INGRESS RancherNetworkExposureType = "INGRESS"
	// gateway.
	RancherNetworkExposureType_GATEWAY RancherNetworkExposureType = "GATEWAY"
	// none.
	RancherNetworkExposureType_NONE RancherNetworkExposureType = "NONE"
)

