package headlamp


type HeadlampIngressTls struct {
	Hosts *[]*string `field:"optional" json:"hosts" yaml:"hosts"`
	SecretName *string `field:"optional" json:"secretName" yaml:"secretName"`
}

