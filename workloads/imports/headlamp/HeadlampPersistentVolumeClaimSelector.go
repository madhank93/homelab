package headlamp


type HeadlampPersistentVolumeClaimSelector struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	MatchExpressions *[]*HeadlampPersistentVolumeClaimSelectorMatchExpressions `field:"optional" json:"matchExpressions" yaml:"matchExpressions"`
	MatchLabels interface{} `field:"optional" json:"matchLabels" yaml:"matchLabels"`
}

