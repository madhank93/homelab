package headlamp


type HeadlampIngressHostsPaths struct {
	Path *string `field:"required" json:"path" yaml:"path"`
	Type *string `field:"required" json:"type" yaml:"type"`
	// Optional override of the backend Service for this path.
	Backend *HeadlampIngressHostsPathsBackend `field:"optional" json:"backend" yaml:"backend"`
}

