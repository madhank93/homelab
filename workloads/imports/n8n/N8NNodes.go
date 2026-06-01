package n8n


// Node configurations for built-in and external npm packages.
type N8NNodes struct {
	// Image for the init container to install npm packages.
	InitContainer *N8NNodesInitContainer `field:"required" json:"initContainer" yaml:"initContainer"`
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	// Enable built-in node functions (e.g., HTTP Request, Code Node, etc.).
	Builtin *N8NNodesBuiltin `field:"optional" json:"builtin" yaml:"builtin"`
	// External npm packages to install and allow in the Code node.
	External *N8NNodesExternal `field:"optional" json:"external" yaml:"external"`
}

