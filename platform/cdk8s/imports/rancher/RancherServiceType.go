package rancher


type RancherServiceType string

const (
	// ClusterIP.
	RancherServiceType_CLUSTER_IP RancherServiceType = "CLUSTER_IP"
	// LoadBalancer.
	RancherServiceType_LOAD_BALANCER RancherServiceType = "LOAD_BALANCER"
	// NodePort.
	RancherServiceType_NODE_PORT RancherServiceType = "NODE_PORT"
)

