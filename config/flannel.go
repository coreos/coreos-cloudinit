package config

type Flannel struct {
	EtcdEndpoint string `yaml:"etcd-endpoint" env:"FLANNELD_ETCD_ENDPOINT"`
	EtcdPrefix   string `yaml:"etcd-prefix"   env:"FLANNELD_ETCD_PREFIX"`
	IPMasq       string `yaml:"ip-masq"       env:"FLANNELD_IP_MASQ"`
	SubnetFile   string `yaml:"subnet-file"   env:"FLANNELD_SUBNET_FILE"`
	Iface        string `yaml:"interface"     env:"FLANNELD_IFACE"`
}
