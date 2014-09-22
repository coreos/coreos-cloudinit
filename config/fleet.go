package config

type Fleet struct {
	AgentTTL                string `yaml:"agent-ttl"                 env:"FLEET_AGENT_TTL"`
	EngineReconcileInterval string `yaml:"engine-reconcile-interval" env:"FLEET_ENGINE_RECONCILE_INTERVAL"`
	EtcdCAFile              string `yaml:"etcd-cafile"               env:"FLEET_ETCD_CAFILE"`
	EtcdCertFile            string `yaml:"etcd-certfile"             env:"FLEET_ETCD_CERTFILE"`
	EtcdKeyFile             string `yaml:"etcd-keyfile"              env:"FLEET_ETCD_KEYFILE"`
	EtcdRequestTimeout      string `yaml:"etcd-request-timeout"      env:"FLEET_ETCD_REQUEST_TIMEOUT"`
	EtcdServers             string `yaml:"etcd-servers"              env:"FLEET_ETCD_SERVERS"`
	Metadata                string `yaml:"metadata"                  env:"FLEET_METADATA"`
	PublicIP                string `yaml:"public-ip"                 env:"FLEET_PUBLIC_IP"`
	Verbosity               string `yaml:"verbosity"                 env:"FLEET_VERBOSITY"`
}
