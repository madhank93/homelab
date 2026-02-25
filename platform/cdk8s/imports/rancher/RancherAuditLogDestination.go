package rancher


// auditLog.destination must be either 'sidecar' or 'hostPath'.
type RancherAuditLogDestination string

const (
	// sidecar.
	RancherAuditLogDestination_SIDECAR RancherAuditLogDestination = "SIDECAR"
	// hostPath.
	RancherAuditLogDestination_HOST_PATH RancherAuditLogDestination = "HOST_PATH"
)

