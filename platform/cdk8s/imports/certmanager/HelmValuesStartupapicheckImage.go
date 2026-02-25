package certmanager


type HelmValuesStartupapicheckImage struct {
	Digest *string `field:"optional" json:"digest" yaml:"digest"`
	PullPolicy *string `field:"optional" json:"pullPolicy" yaml:"pullPolicy"`
	Registry *string `field:"optional" json:"registry" yaml:"registry"`
	Repository *string `field:"optional" json:"repository" yaml:"repository"`
	Tag *string `field:"optional" json:"tag" yaml:"tag"`
}

