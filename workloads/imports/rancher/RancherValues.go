package rancher


type RancherValues struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// agentTLSMode must be 'strict' or 'system-store' or null (defaults to system-store).
	AgentTlsMode RancherAgentTlsMode `field:"optional" json:"agentTlsMode" yaml:"agentTlsMode"`
	AuditLog *RancherAuditLog `field:"optional" json:"auditLog" yaml:"auditLog"`
	Global *map[string]interface{} `field:"optional" json:"global" yaml:"global"`
	// The default rancher ingress configuration.
	Ingress *RancherIngress `field:"optional" json:"ingress" yaml:"ingress"`
	// The default rancher service configuration.
	Service *RancherService `field:"optional" json:"service" yaml:"service"`
}

