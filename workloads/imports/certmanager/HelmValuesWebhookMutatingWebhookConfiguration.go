package certmanager


type HelmValuesWebhookMutatingWebhookConfiguration struct {
	NamespaceSelector interface{} `field:"optional" json:"namespaceSelector" yaml:"namespaceSelector"`
}

