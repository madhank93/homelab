package n8n


// External Redis parameters.
type N8NExternalRedis struct {
	// List of Redis Cluster nodes.
	//
	// Setting this variable will create a Redis Cluster client instead of a Redis client, and n8n will ignore `externalRedis.host` and `externalRedis.port`.
	ClusterNodes *[]*N8NExternalRedisClusterNodes `field:"required" json:"clusterNodes" yaml:"clusterNodes"`
	// Redis database for Bull queue.
	Database *float64 `field:"required" json:"database" yaml:"database"`
	// Enable dual-stack support (IPv4 and IPv6) on Redis connections.
	DualStack *bool `field:"required" json:"dualStack" yaml:"dualStack"`
	// The name of an existing secret with Redis (must contain key `redis-password`) and Sentinel credentials.
	//
	// When it's set, the `externalRedis.password` parameter is ignored
	ExistingSecret *string `field:"required" json:"existingSecret" yaml:"existingSecret"`
	// External Redis server host.
	Host *string `field:"required" json:"host" yaml:"host"`
	// External Redis password.
	Password *string `field:"required" json:"password" yaml:"password"`
	// External Redis server port.
	Port *float64 `field:"required" json:"port" yaml:"port"`
	// Placeholder for future Redis TLS certificates.
	Tls *N8NExternalRedisTls `field:"required" json:"tls" yaml:"tls"`
	// External Redis username.
	Username *string `field:"required" json:"username" yaml:"username"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

