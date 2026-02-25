package n8n


// This sets the container image more information can be found here: https://kubernetes.io/docs/concepts/containers/images/.
type N8NImage struct {
	// This sets the pull policy for images.
	PullPolicy *string `field:"required" json:"pullPolicy" yaml:"pullPolicy"`
	Repository *string `field:"required" json:"repository" yaml:"repository"`
	// Overrides the image tag whose default is the chart appVersion.
	Tag *string `field:"required" json:"tag" yaml:"tag"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

