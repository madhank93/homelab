package n8n


// Enable built-in node functions (e.g., HTTP Request, Code Node, etc.).
type N8NNodesBuiltin struct {
	// Enable built-in modules for the Code node.
	Enabled *bool `field:"optional" json:"enabled" yaml:"enabled"`
	// List of built-in Node.js modules to allow in the Code node (e.g., crypto, fs). Use '*' to allow all.
	Modules *[]*string `field:"optional" json:"modules" yaml:"modules"`
}

