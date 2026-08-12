package headlamp


// Selects a field of the pod.
type HeadlampEnvValueFromFieldRef struct {
	// Path of the field to select in the specified API version.
	FieldPath *string `field:"required" json:"fieldPath" yaml:"fieldPath"`
	// API version of the schema the fieldPath is written in.
	ApiVersion *string `field:"optional" json:"apiVersion" yaml:"apiVersion"`
}

