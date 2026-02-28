package n8n


type N8NWorker struct {
	// For more information checkout: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity.
	Affinity interface{} `field:"required" json:"affinity" yaml:"affinity"`
	// If true, all k8s nodes will deploy exatly one worker pod.
	AllNodes *bool `field:"required" json:"allNodes" yaml:"allNodes"`
	Autoscaling *N8NWorkerAutoscaling `field:"required" json:"autoscaling" yaml:"autoscaling"`
	// number of concurrency for each worker.
	Concurrency *float64 `field:"required" json:"concurrency" yaml:"concurrency"`
	// number of workers.
	Count *float64 `field:"required" json:"count" yaml:"count"`
	// Additional containers for the worker pod.
	ExtraContainers *[]interface{} `field:"required" json:"extraContainers" yaml:"extraContainers"`
	// Extra environment variables.
	ExtraEnvVars interface{} `field:"required" json:"extraEnvVars" yaml:"extraEnvVars"`
	// Extra secrets for environment variables.
	ExtraSecretNamesForEnvFrom *[]interface{} `field:"required" json:"extraSecretNamesForEnvFrom" yaml:"extraSecretNamesForEnvFrom"`
	// Force to use statefulset for the worker pod.
	//
	// If true, the worker pod will be created as a statefulset.
	ForceToUseStatefulset *bool `field:"required" json:"forceToUseStatefulset" yaml:"forceToUseStatefulset"`
	// Host aliases for the worker pod.
	//
	// For more information checkout: https://kubernetes.io/docs/tasks/network/customize-hosts-file-for-pods/#adding-additional-entries-with-hostaliases
	HostAliases *[]*N8NWorkerHostAliases `field:"required" json:"hostAliases" yaml:"hostAliases"`
	// Additional init containers for the worker pod.
	InitContainers *[]interface{} `field:"required" json:"initContainers" yaml:"initContainers"`
	// This is to setup the liveness probe for the worker pod more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/.
	LivenessProbe *N8NWorkerLivenessProbe `field:"required" json:"livenessProbe" yaml:"livenessProbe"`
	// Use `regular` to use main node as executer, or use `queue` to have worker nodes.
	Mode N8NWorkerMode `field:"required" json:"mode" yaml:"mode"`
	// PodDisruptionBudget configuration for the worker node.
	Pdb *N8NWorkerPdb `field:"required" json:"pdb" yaml:"pdb"`
	// Persistence configuration for the worker pod.
	Persistence *N8NWorkerPersistence `field:"required" json:"persistence" yaml:"persistence"`
	// This is to setup the readiness probe for the worker pod more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/.
	ReadinessProbe *N8NWorkerReadinessProbe `field:"required" json:"readinessProbe" yaml:"readinessProbe"`
	// This block is for setting up the resource management for the pod more information can be found here: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
	Resources *N8NWorkerResources `field:"required" json:"resources" yaml:"resources"`
	// Runtime class name for the worker pod.
	//
	// For more information checkout: https://kubernetes.io/docs/concepts/containers/runtime-class/
	RuntimeClassName *string `field:"required" json:"runtimeClassName" yaml:"runtimeClassName"`
	// This is to setup the startup probe for the worker pod more information can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/.
	StartupProbe *N8NWorkerStartupProbe `field:"required" json:"startupProbe" yaml:"startupProbe"`
	// Additional volumeMounts on the output Deployment definition.
	VolumeMounts *[]interface{} `field:"required" json:"volumeMounts" yaml:"volumeMounts"`
	// Additional volumes on the output Deployment definition.
	Volumes *[]interface{} `field:"required" json:"volumes" yaml:"volumes"`
	// This is to setup the wait for the main node to be ready.
	WaitMainNodeReady *N8NWorkerWaitMainNodeReady `field:"required" json:"waitMainNodeReady" yaml:"waitMainNodeReady"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
}

