package n8n


// Redis Cluster node.
type N8NExternalRedisClusterNodes struct {
	// External Redis server host.
	Host *string `field:"required" json:"host" yaml:"host"`
	// External Redis server port.
	Port *float64 `field:"required" json:"port" yaml:"port"`
}

