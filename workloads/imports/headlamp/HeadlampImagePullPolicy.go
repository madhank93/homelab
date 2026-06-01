package headlamp


// Pull policy of the image.
type HeadlampImagePullPolicy string

const (
	// Always.
	HeadlampImagePullPolicy_ALWAYS HeadlampImagePullPolicy = "ALWAYS"
	// IfNotPresent.
	HeadlampImagePullPolicy_IF_NOT_PRESENT HeadlampImagePullPolicy = "IF_NOT_PRESENT"
	// Never.
	HeadlampImagePullPolicy_NEVER HeadlampImagePullPolicy = "NEVER"
)

