package certmanager


type HelmValuesPrometheus struct {
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	Podmonitor *HelmValuesPrometheusPodmonitor `field:"optional" json:"podmonitor" yaml:"podmonitor"`
	Servicemonitor *HelmValuesPrometheusServicemonitor `field:"optional" json:"servicemonitor" yaml:"servicemonitor"`
}

