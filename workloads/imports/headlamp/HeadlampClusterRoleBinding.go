package headlamp


type HeadlampClusterRoleBinding struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Annotations to add to the cluster role binding.
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	// The name of the ClusterRole to create in the cluster.
	ClusterRoleName *string `field:"optional" json:"clusterRoleName" yaml:"clusterRoleName"`
	// Specifies whether a cluster role binding should be created.
	Create *bool `field:"optional" json:"create" yaml:"create"`
}

