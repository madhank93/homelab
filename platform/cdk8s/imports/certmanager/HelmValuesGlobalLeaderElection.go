package certmanager


type HelmValuesGlobalLeaderElection struct {
	LeaseDuration *string `field:"optional" json:"leaseDuration" yaml:"leaseDuration"`
	Namespace *string `field:"optional" json:"namespace" yaml:"namespace"`
	RenewDeadline *string `field:"optional" json:"renewDeadline" yaml:"renewDeadline"`
	RetryPeriod *string `field:"optional" json:"retryPeriod" yaml:"retryPeriod"`
}

