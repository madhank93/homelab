package rancher


type RancherProps struct {
	HelmExecutable *string `field:"optional" json:"helmExecutable" yaml:"helmExecutable"`
	HelmFlags *[]*string `field:"optional" json:"helmFlags" yaml:"helmFlags"`
	Namespace *string `field:"optional" json:"namespace" yaml:"namespace"`
	ReleaseName *string `field:"optional" json:"releaseName" yaml:"releaseName"`
	Values *RancherValues `field:"optional" json:"values" yaml:"values"`
}

