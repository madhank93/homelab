package grafana


// BETA: Configure Gateway API ListenerSet resources for the chart here.
//
// ListenerSet allows attaching additional listeners to an existing Gateway. More listener sets can be added by adding a dictionary key like the 'main' entry. Being BETA this can/will change in the future without notice, do not use unless you want to take that risk [[ref]](https://gateway-api.sigs.k8s.io/reference/api-spec/main/spec/#listenerset)
type GrafanaListenerSet struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Main *GrafanaListenerSetMain `field:"optional" json:"main" yaml:"main"`
}

