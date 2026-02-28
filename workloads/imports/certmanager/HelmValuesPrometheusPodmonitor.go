package certmanager


type HelmValuesPrometheusPodmonitor struct {
	Annotations interface{} `field:"optional" json:"annotations" yaml:"annotations"`
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	EndpointAdditionalProperties interface{} `field:"optional" json:"endpointAdditionalProperties" yaml:"endpointAdditionalProperties"`
	HonorLabels *bool `field:"optional" json:"honorLabels" yaml:"honorLabels"`
	Interval *string `field:"optional" json:"interval" yaml:"interval"`
	Labels interface{} `field:"optional" json:"labels" yaml:"labels"`
	Namespace *string `field:"optional" json:"namespace" yaml:"namespace"`
	Path *string `field:"optional" json:"path" yaml:"path"`
	PrometheusInstance *string `field:"optional" json:"prometheusInstance" yaml:"prometheusInstance"`
	ScrapeTimeout *string `field:"optional" json:"scrapeTimeout" yaml:"scrapeTimeout"`
}

