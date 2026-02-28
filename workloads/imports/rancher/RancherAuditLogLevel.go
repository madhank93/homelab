package rancher


// auditLog.level must be a number 0-3; 0 to disable, 3 for most verbose.
type RancherAuditLogLevel string

const (
	// 0.
	RancherAuditLogLevel_VALUE_0 RancherAuditLogLevel = "VALUE_0"
	// 1.
	RancherAuditLogLevel_VALUE_1 RancherAuditLogLevel = "VALUE_1"
	// 2.
	RancherAuditLogLevel_VALUE_2 RancherAuditLogLevel = "VALUE_2"
	// 3.
	RancherAuditLogLevel_VALUE_3 RancherAuditLogLevel = "VALUE_3"
)

