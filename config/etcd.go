package config

type Etcd struct {
	Addr                string `yaml:"addr"                  env:"ETCD_ADDR"`
	BindAddr            string `yaml:"bind-addr"             env:"ETCD_BIND_ADDR"`
	CAFile              string `yaml:"ca-file"               env:"ETCD_CA_FILE"`
	CertFile            string `yaml:"cert-file"             env:"ETCD_CERT_FILE"`
	ClusterActiveSize   string `yaml:"cluster-active-size"   env:"ETCD_CLUSTER_ACTIVE_SIZE"`
	ClusterRemoveDelay  string `yaml:"cluster-remove-delay"  env:"ETCD_CLUSTER_REMOVE_DELAY"`
	ClusterSyncInterval string `yaml:"cluster-sync-interval" env:"ETCD_CLUSTER_SYNC_INTERVAL"`
	Cors                string `yaml:"cors"                  env:"ETCD_CORS"`
	CPUProfileFile      string `yaml:"cpu-profile-file"      env:"ETCD_CPU_PROFILE_FILE"`
	DataDir             string `yaml:"data-dir"              env:"ETCD_DATA_DIR"`
	Discovery           string `yaml:"discovery"             env:"ETCD_DISCOVERY"`
	HTTPReadTimeout     string `yaml:"http-read-timeout"     env:"ETCD_HTTP_READ_TIMEOUT"`
	HTTPWriteTimeout    string `yaml:"http-write-timeout"    env:"ETCD_HTTP_WRITE_TIMEOUT"`
	KeyFile             string `yaml:"key-file"              env:"ETCD_KEY_FILE"`
	MaxClusterSize      string `yaml:"max-cluster-size"      env:"ETCD_MAX_CLUSTER_SIZE"`
	MaxResultBuffer     string `yaml:"max-result-buffer"     env:"ETCD_MAX_RESULT_BUFFER"`
	MaxRetryAttempts    string `yaml:"max-retry-attempts"    env:"ETCD_MAX_RETRY_ATTEMPTS"`
	Name                string `yaml:"name"                  env:"ETCD_NAME"`
	PeerAddr            string `yaml:"peer-addr"             env:"ETCD_PEER_ADDR"`
	PeerBindAddr        string `yaml:"peer-bind-addr"        env:"ETCD_PEER_BIND_ADDR"`
	PeerCAFile          string `yaml:"peer-ca-file"          env:"ETCD_PEER_CA_FILE"`
	PeerCertFile        string `yaml:"peer-cert-file"        env:"ETCD_PEER_CERT_FILE"`
	PeerKeyFile         string `yaml:"peer-key-file"         env:"ETCD_PEER_KEY_FILE"`
	Peers               string `yaml:"peers"                 env:"ETCD_PEERS"`
	PeersFile           string `yaml:"peers-file"            env:"ETCD_PEERS_FILE"`
	Snapshot            string `yaml:"snapshot"              env:"ETCD_SNAPSHOT"`
	Verbose             string `yaml:"verbose"               env:"ETCD_VERBOSE"`
	VeryVerbose         string `yaml:"very-verbose"          env:"ETCD_VERY_VERBOSE"`
}
