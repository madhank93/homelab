package headlamp


// Image to deploy.
type HeadlampImage struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Pull policy of the image.
	PullPolicy HeadlampImagePullPolicy `field:"optional" json:"pullPolicy" yaml:"pullPolicy"`
	// Registry of the image.
	Registry *string `field:"optional" json:"registry" yaml:"registry"`
	// Repository of the image.
	Repository *string `field:"optional" json:"repository" yaml:"repository"`
	// Tag of the image.
	Tag *string `field:"optional" json:"tag" yaml:"tag"`
}

