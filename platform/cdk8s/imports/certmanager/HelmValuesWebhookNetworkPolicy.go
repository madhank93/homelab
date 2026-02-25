package certmanager


type HelmValuesWebhookNetworkPolicy struct {
	Egress *[]interface{} `field:"optional" json:"egress" yaml:"egress"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Ingress *[]interface{} `field:"optional" json:"ingress" yaml:"ingress"`
}

