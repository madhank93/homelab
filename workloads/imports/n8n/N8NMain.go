package n8n


// Main node configuration.
type N8NMain struct {
	// For more information checkout: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity.
	Affinity interface{} `field:"required" json:"affinity" yaml:"affinity"`
	// Number of main nodes.
	//
	// Only enterprise license users can have one leader main node and mutiple follower main nodes.
	Count *float64 `field:"required" json:"count" yaml:"count"`
	// Editor based URL.
	//
	// If it's not defined and ingress definition exists, ingress host will be used.
	EditorBaseUrl *string `field:"required" json:"editorBaseUrl" yaml:"editorBaseUrl"`
	// Additional containers for the main pod.
	ExtraContainers *[]interface{} `field:"required" json:"extraContainers" yaml:"extraContainers"`
	// Extra environment variables.
	ExtraEnvVars interface{} `field:"required" json:"extraEnvVars" yaml:"extraEnvVars"`
	// Extra secrets for environment variables.
	ExtraSecretNamesForEnvFrom *[]interface{} `field:"required" json:"extraSecretNamesForEnvFrom" yaml:"extraSecretNamesForEnvFrom"`
	// Force to use statefulset for the main pod.
	//
	// If true, the main pod will be created as a statefulset.
	ForceToUseStatefulset *bool `field:"required" json:"forceToUseStatefulset" yaml:"forceToUseStatefulset"`
	// Host aliases for the main pod.
	//
	// For more information checkout: https://kubernetes.io/docs/tasks/network/customize-hosts-file-for-pods/#adding-additional-entries-with-hostaliases
	HostAliases *[]*N8NMainHostAliases `field:"required" json:"hostAliases" yaml:"hostAliases"`
	// Additional init containers for the main pod.
	InitContainers *[]interface{} `field:"required" json:"initContainers" yaml:"initContainers"`
	// This is to setup the liveness probe for the main pod more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/.
	LivenessProbe *N8NMainLivenessProbe `field:"required" json:"livenessProbe" yaml:"livenessProbe"`
	// PodDisruptionBudget configuration for the main node.
	Pdb *N8NMainPdb `field:"required" json:"pdb" yaml:"pdb"`
	// Persistence configuration for the main pod.
	Persistence *N8NMainPersistence `field:"required" json:"persistence" yaml:"persistence"`
	// This is to setup the readiness probe for the main pod more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/.
	ReadinessProbe *N8NMainReadinessProbe `field:"required" json:"readinessProbe" yaml:"readinessProbe"`
	// This block is for setting up the resource management for the pod more information can be found here: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
	Resources *N8NMainResources `field:"required" json:"resources" yaml:"resources"`
	// Runtime class name for the main pod.
	//
	// For more information checkout: https://kubernetes.io/docs/concepts/containers/runtime-class/
	RuntimeClassName *string `field:"required" json:"runtimeClassName" yaml:"runtimeClassName"`
	// Additional volumeMounts on the output Deployment definition.
	VolumeMounts *[]interface{} `field:"required" json:"volumeMounts" yaml:"volumeMounts"`
	// Additional volumes on the output Deployment definition.
	Volumes *[]interface{} `field:"required" json:"volumes" yaml:"volumes"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

