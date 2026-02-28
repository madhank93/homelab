package headlamp


// Resources of the init container.
type HeadlampInitContainersResources struct {
	// Limits of the init container.
	Limits *HeadlampInitContainersResourcesLimits `field:"optional" json:"limits" yaml:"limits"`
	// Requests of the init container.
	Requests *HeadlampInitContainersResourcesRequests `field:"optional" json:"requests" yaml:"requests"`
}

