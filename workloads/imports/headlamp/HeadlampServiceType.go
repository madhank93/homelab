package headlamp


// Kubernetes Service type.
type HeadlampServiceType string

const (
	// ClusterIP.
	HeadlampServiceType_CLUSTER_IP HeadlampServiceType = "CLUSTER_IP"
	// NodePort.
	HeadlampServiceType_NODE_PORT HeadlampServiceType = "NODE_PORT"
	// LoadBalancer.
	HeadlampServiceType_LOAD_BALANCER HeadlampServiceType = "LOAD_BALANCER"
	// ExternalName.
	HeadlampServiceType_EXTERNAL_NAME HeadlampServiceType = "EXTERNAL_NAME"
)

