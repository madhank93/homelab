package n8n


// A DNS option object.
type N8NDnsConfigOptions struct {
	// Name of the DNS option.
	Name *string `field:"required" json:"name" yaml:"name"`
	// Value of the DNS option (optional).
	Value *string `field:"optional" json:"value" yaml:"value"`
}

