package headlamp


type HeadlampHttpRouteParentRefs struct {
	// Name of the parent gateway.
	Name *string `field:"required" json:"name" yaml:"name"`
	// Namespace of the parent gateway.
	Namespace *string `field:"optional" json:"namespace" yaml:"namespace"`
	// Section name of the parent gateway listener.
	SectionName *string `field:"optional" json:"sectionName" yaml:"sectionName"`
}

