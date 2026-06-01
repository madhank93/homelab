package headlamp


// Pull policy of the init container.
type HeadlampInitContainersImagePullPolicy string

const (
	// Always.
	HeadlampInitContainersImagePullPolicy_ALWAYS HeadlampInitContainersImagePullPolicy = "ALWAYS"
	// IfNotPresent.
	HeadlampInitContainersImagePullPolicy_IF_NOT_PRESENT HeadlampInitContainersImagePullPolicy = "IF_NOT_PRESENT"
	// Never.
	HeadlampInitContainersImagePullPolicy_NEVER HeadlampInitContainersImagePullPolicy = "NEVER"
)

