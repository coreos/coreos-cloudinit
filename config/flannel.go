package config

type Flannel struct {
	EtcdEndpoint string `yaml:"etcd_endpoint" env:"FLANNELD_ETCD_ENDPOINT"`
	EtcdPrefix   string `yaml:"etcd_prefix"   env:"FLANNELD_ETCD_PREFIX"`
	IPMasq       string `yaml:"ip_masq"       env:"FLANNELD_IP_MASQ"`
	SubnetFile   string `yaml:"subnet_file"   env:"FLANNELD_SUBNET_FILE"`
	Iface        string `yaml:"interface"     env:"FLANNELD_IFACE"`
}
