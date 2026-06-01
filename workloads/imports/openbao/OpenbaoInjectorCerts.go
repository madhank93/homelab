package openbao


type OpenbaoInjectorCerts struct {
	// Values that are not available in values.schema.json will not be code generated. You can add such values to this property.
	AdditionalValues *map[string]interface{} `field:"optional" json:"additionalValues" yaml:"additionalValues"`
	CaBundle *string `field:"optional" json:"caBundle" yaml:"caBundle"`
	CertName *string `field:"optional" json:"certName" yaml:"certName"`
	KeyName *string `field:"optional" json:"keyName" yaml:"keyName"`
	SecretName *string `field:"optional" json:"secretName" yaml:"secretName"`
}

