package certmanager


type HelmValuesGlobalRbac struct {
	AggregateClusterRoles *bool `field:"optional" json:"aggregateClusterRoles" yaml:"aggregateClusterRoles"`
	Create *bool `field:"optional" json:"create" yaml:"create"`
}

