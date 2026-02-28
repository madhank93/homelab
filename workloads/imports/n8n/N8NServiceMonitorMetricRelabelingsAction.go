package n8n


// The relabeling action to perform.
type N8NServiceMonitorMetricRelabelingsAction string

const (
	// replace.
	N8NServiceMonitorMetricRelabelingsAction_REPLACE N8NServiceMonitorMetricRelabelingsAction = "REPLACE"
	// keep.
	N8NServiceMonitorMetricRelabelingsAction_KEEP N8NServiceMonitorMetricRelabelingsAction = "KEEP"
	// drop.
	N8NServiceMonitorMetricRelabelingsAction_DROP N8NServiceMonitorMetricRelabelingsAction = "DROP"
	// labeldrop.
	N8NServiceMonitorMetricRelabelingsAction_LABELDROP N8NServiceMonitorMetricRelabelingsAction = "LABELDROP"
	// labelkeep.
	N8NServiceMonitorMetricRelabelingsAction_LABELKEEP N8NServiceMonitorMetricRelabelingsAction = "LABELKEEP"
	// hashmod.
	N8NServiceMonitorMetricRelabelingsAction_HASHMOD N8NServiceMonitorMetricRelabelingsAction = "HASHMOD"
)

