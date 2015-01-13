package config

type Flannel struct {
	EtcdEndpoints string `yaml:"etcd_endpoints" env:"FLANNELD_ETCD_ENDPOINTS"`
	EtcdCAFile    string `yaml:"etcd_cafile"    env:"FLANNELD_ETCD_CAFILE"`
	EtcdCertFile  string `yaml:"etcd_certfile"  env:"FLANNELD_ETCD_CERTFILE"`
	EtcdKeyFile   string `yaml:"etcd_keyfile"   env:"FLANNELD_ETCD_KEYFILE"`
	EtcdPrefix    string `yaml:"etcd_prefix"    env:"FLANNELD_ETCD_PREFIX"`
	IPMasq        string `yaml:"ip_masq"        env:"FLANNELD_IP_MASQ"`
	SubnetFile    string `yaml:"subnet_file"    env:"FLANNELD_SUBNET_FILE"`
	Iface         string `yaml:"interface"      env:"FLANNELD_IFACE"`
}
