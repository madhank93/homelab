package certmanager


type HelmValuesCrds struct {
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Keep *bool `field:"optional" json:"keep" yaml:"keep"`
}

