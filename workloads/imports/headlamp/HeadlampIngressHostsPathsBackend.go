package headlamp


// Optional override of the backend Service for this path.
type HeadlampIngressHostsPathsBackend struct {
	Service *HeadlampIngressHostsPathsBackendService `field:"optional" json:"service" yaml:"service"`
}

