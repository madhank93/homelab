package headlamp


type HeadlampIngressHosts struct {
	Host *string `field:"optional" json:"host" yaml:"host"`
	Paths *[]*HeadlampIngressHostsPaths `field:"optional" json:"paths" yaml:"paths"`
}

