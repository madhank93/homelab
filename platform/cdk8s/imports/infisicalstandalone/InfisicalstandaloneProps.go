package infisicalstandalone


type InfisicalstandaloneProps struct {
	HelmExecutable *string `field:"optional" json:"helmExecutable" yaml:"helmExecutable"`
	HelmFlags *[]*string `field:"optional" json:"helmFlags" yaml:"helmFlags"`
	Namespace *string `field:"optional" json:"namespace" yaml:"namespace"`
	ReleaseName *string `field:"optional" json:"releaseName" yaml:"releaseName"`
	Values *map[string]interface{} `field:"optional" json:"values" yaml:"values"`
}

