package n8n


// This is for setting Security Context to a Pod.
//
// For more information checkout: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
type N8NPodSecurityContext struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	FsGroup *float64 `field:"optional" json:"fsGroup" yaml:"fsGroup"`
	FsGroupChangePolicy N8NPodSecurityContextFsGroupChangePolicy `field:"optional" json:"fsGroupChangePolicy" yaml:"fsGroupChangePolicy"`
	RunAsGroup *float64 `field:"optional" json:"runAsGroup" yaml:"runAsGroup"`
	RunAsUser *float64 `field:"optional" json:"runAsUser" yaml:"runAsUser"`
	SupplementalGroups *[]interface{} `field:"optional" json:"supplementalGroups" yaml:"supplementalGroups"`
}

