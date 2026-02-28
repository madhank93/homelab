package headlamp


type HeadlampPersistentVolumeClaimSelectorMatchExpressions struct {
	Key *string `field:"optional" json:"key" yaml:"key"`
	Operator *string `field:"optional" json:"operator" yaml:"operator"`
	Values *[]*string `field:"optional" json:"values" yaml:"values"`
}

