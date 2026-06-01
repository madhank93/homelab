package certmanager


type HelmValuesGlobalPodSecurityPolicy struct {
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	UseAppArmor *bool `field:"optional" json:"useAppArmor" yaml:"useAppArmor"`
}

