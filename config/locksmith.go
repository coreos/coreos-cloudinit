package config

type Locksmith struct {
	Endpoint     string `yaml:"endpoint"      env:"LOCKSMITHD_ENDPOINT"`
	EtcdCAFile   string `yaml:"etcd_cafile"   env:"LOCKSMITHD_ETCD_CAFILE"`
	EtcdCertFile string `yaml:"etcd_certfile" env:"LOCKSMITHD_ETCD_CERTFILE"`
	EtcdKeyFile  string `yaml:"etcd_keyfile"  env:"LOCKSMITHD_ETCD_KEYFILE"`
}
