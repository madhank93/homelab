package secretsstorecsixk8sio


// SecretObjectData defines the desired state of synced K8s secret object data.
type SecretProviderClassSpecSecretObjectsData struct {
	// data field to populate.
	Key *string `field:"optional" json:"key" yaml:"key"`
	// name of the object to sync.
	ObjectName *string `field:"optional" json:"objectName" yaml:"objectName"`
}

