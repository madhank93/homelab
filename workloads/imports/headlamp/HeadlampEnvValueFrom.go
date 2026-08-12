package headlamp


// Source for the environment variable's value.
type HeadlampEnvValueFrom struct {
	// Selects a key of a ConfigMap.
	ConfigMapKeyRef *HeadlampEnvValueFromConfigMapKeyRef `field:"optional" json:"configMapKeyRef" yaml:"configMapKeyRef"`
	// Selects a field of the pod.
	FieldRef *HeadlampEnvValueFromFieldRef `field:"optional" json:"fieldRef" yaml:"fieldRef"`
	// Selects a resource of the container.
	ResourceFieldRef *HeadlampEnvValueFromResourceFieldRef `field:"optional" json:"resourceFieldRef" yaml:"resourceFieldRef"`
	// Selects a key of a Secret.
	SecretKeyRef *HeadlampEnvValueFromSecretKeyRef `field:"optional" json:"secretKeyRef" yaml:"secretKeyRef"`
}

