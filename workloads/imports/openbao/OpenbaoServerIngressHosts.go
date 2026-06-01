package openbao


type OpenbaoServerIngressHosts struct {
	Host *string `field:"optional" json:"host" yaml:"host"`
	Paths *[]interface{} `field:"optional" json:"paths" yaml:"paths"`
}

