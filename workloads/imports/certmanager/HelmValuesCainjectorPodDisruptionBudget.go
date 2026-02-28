package certmanager


type HelmValuesCainjectorPodDisruptionBudget struct {
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	MaxUnavailable interface{} `field:"optional" json:"maxUnavailable" yaml:"maxUnavailable"`
	MinAvailable interface{} `field:"optional" json:"minAvailable" yaml:"minAvailable"`
}

