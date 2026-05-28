package rancher


type RancherGatewayGatewayClassTlsSource string

const (
	// rancher.
	RancherGatewayGatewayClassTlsSource_RANCHER RancherGatewayGatewayClassTlsSource = "RANCHER"
	// letsEncrypt.
	RancherGatewayGatewayClassTlsSource_LETS_ENCRYPT RancherGatewayGatewayClassTlsSource = "LETS_ENCRYPT"
	// secret.
	RancherGatewayGatewayClassTlsSource_SECRET RancherGatewayGatewayClassTlsSource = "SECRET"
)

