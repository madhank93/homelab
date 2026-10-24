package grafana


// BETA: Configure the gateway routes for the chart here.
//
// More routes can be added by adding a dictionary key like the 'main' route. Be aware that this is an early beta of this feature, kube-prometheus-stack does not guarantee this works and is subject to change. Being BETA this can/will change in the future without notice, do not use unless you want to take that risk [[ref]](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io%2fv1alpha2)
type GrafanaRoute struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	Main *GrafanaRouteMain `field:"optional" json:"main" yaml:"main"`
}

