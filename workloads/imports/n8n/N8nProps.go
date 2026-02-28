package n8n


type N8nProps struct {
	Values *N8nValues `field:"required" json:"values" yaml:"values"`
	HelmExecutable *string `field:"optional" json:"helmExecutable" yaml:"helmExecutable"`
	HelmFlags *[]*string `field:"optional" json:"helmFlags" yaml:"helmFlags"`
	Namespace *string `field:"optional" json:"namespace" yaml:"namespace"`
	ReleaseName *string `field:"optional" json:"releaseName" yaml:"releaseName"`
}

