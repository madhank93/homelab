package rancher


// agentTLSMode must be 'strict' or 'system-store' or null (defaults to system-store).
type RancherAgentTlsMode string

const (
	// strict.
	RancherAgentTlsMode_STRICT RancherAgentTlsMode = "STRICT"
	// system-store.
	RancherAgentTlsMode_SYSTEM_HYPHEN_STORE RancherAgentTlsMode = "SYSTEM_HYPHEN_STORE"
)

