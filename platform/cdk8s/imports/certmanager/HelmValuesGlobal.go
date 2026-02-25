package certmanager


// Global values shared across all (sub)charts.
type HelmValuesGlobal struct {
	CommonLabels interface{} `field:"optional" json:"commonLabels" yaml:"commonLabels"`
	HostUsers *bool `field:"optional" json:"hostUsers" yaml:"hostUsers"`
	ImagePullSecrets *[]interface{} `field:"optional" json:"imagePullSecrets" yaml:"imagePullSecrets"`
	LeaderElection *HelmValuesGlobalLeaderElection `field:"optional" json:"leaderElection" yaml:"leaderElection"`
	LogLevel *float64 `field:"optional" json:"logLevel" yaml:"logLevel"`
	NodeSelector interface{} `field:"optional" json:"nodeSelector" yaml:"nodeSelector"`
	PodSecurityPolicy *HelmValuesGlobalPodSecurityPolicy `field:"optional" json:"podSecurityPolicy" yaml:"podSecurityPolicy"`
	PriorityClassName *string `field:"optional" json:"priorityClassName" yaml:"priorityClassName"`
	Rbac *HelmValuesGlobalRbac `field:"optional" json:"rbac" yaml:"rbac"`
	RevisionHistoryLimit *float64 `field:"optional" json:"revisionHistoryLimit" yaml:"revisionHistoryLimit"`
}

