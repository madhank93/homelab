package certmanager


type HelmValuesIngressShim struct {
	DefaultIssuerGroup *string `field:"optional" json:"defaultIssuerGroup" yaml:"defaultIssuerGroup"`
	DefaultIssuerKind *string `field:"optional" json:"defaultIssuerKind" yaml:"defaultIssuerKind"`
	DefaultIssuerName *string `field:"optional" json:"defaultIssuerName" yaml:"defaultIssuerName"`
}

