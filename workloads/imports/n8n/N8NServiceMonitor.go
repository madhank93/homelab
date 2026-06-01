package n8n


// The ServiceMonitor configuration for the n8n deployment.
//
// Please refer to the following link for more information: https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api-reference/api.md
type N8NServiceMonitor struct {
	// Whether to enable the ServiceMonitor for the n8n deployment.
	Enabled *bool `field:"required" json:"enabled" yaml:"enabled"`
	// Metrics and labels to include in the ServiceMonitor.
	Include *N8NServiceMonitorInclude `field:"required" json:"include" yaml:"include"`
	// The interval for the ServiceMonitor (e.g., 30s, 1m, 1h).
	Interval *string `field:"required" json:"interval" yaml:"interval"`
	// The labels for the ServiceMonitor, use this to define your scrape label for Prometheus Operator.
	Labels *map[string]*string `field:"required" json:"labels" yaml:"labels"`
	// The metric relabelings for the ServiceMonitor, following Prometheus relabel_config structure.
	MetricRelabelings *[]*N8NServiceMonitorMetricRelabelings `field:"required" json:"metricRelabelings" yaml:"metricRelabelings"`
	// The metrics prefix for the ServiceMonitor.
	MetricsPrefix *string `field:"required" json:"metricsPrefix" yaml:"metricsPrefix"`
	// The namespace for the ServiceMonitor.
	//
	// If empty, the ServiceMonitor will be deployed in the same namespace as the n8n chart.
	Namespace *string `field:"required" json:"namespace" yaml:"namespace"`
	// Set of labels to transfer from the Kubernetes Service onto the target.
	TargetLabels *[]*string `field:"required" json:"targetLabels" yaml:"targetLabels"`
	// The timeout for the ServiceMonitor (e.g., 10s, 1m).
	Timeout *string `field:"required" json:"timeout" yaml:"timeout"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

