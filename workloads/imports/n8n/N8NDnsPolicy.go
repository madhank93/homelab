package n8n


// For more information checkout: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy.
type N8NDnsPolicy string

const (
	// ClusterFirst.
	N8NDnsPolicy_CLUSTER_FIRST N8NDnsPolicy = "CLUSTER_FIRST"
	// ClusterFirstWithHostNet.
	N8NDnsPolicy_CLUSTER_FIRST_WITH_HOST_NET N8NDnsPolicy = "CLUSTER_FIRST_WITH_HOST_NET"
	// Default.
	N8NDnsPolicy_DEFAULT N8NDnsPolicy = "DEFAULT"
	// None.
	N8NDnsPolicy_NONE N8NDnsPolicy = "NONE"
)

