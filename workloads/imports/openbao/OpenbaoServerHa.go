package openbao


type OpenbaoServerHa struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	ApiAddr *string `field:"optional" json:"apiAddr" yaml:"apiAddr"`
	ClusterAddr *string `field:"optional" json:"clusterAddr" yaml:"clusterAddr"`
	Config interface{} `field:"optional" json:"config" yaml:"config"`
	DisruptionBudget *OpenbaoServerHaDisruptionBudget `field:"optional" json:"disruptionBudget" yaml:"disruptionBudget"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Raft *OpenbaoServerHaRaft `field:"optional" json:"raft" yaml:"raft"`
	Replicas *float64 `field:"optional" json:"replicas" yaml:"replicas"`
}

