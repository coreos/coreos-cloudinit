/*
   Copyright 2014 CoreOS, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package config

type Etcd struct {
	Addr                  string  `yaml:"addr"                    env:"ETCD_ADDR"`
	BindAddr              string  `yaml:"bind_addr"               env:"ETCD_BIND_ADDR"`
	CAFile                string  `yaml:"ca_file"                 env:"ETCD_CA_FILE"`
	CertFile              string  `yaml:"cert_file"               env:"ETCD_CERT_FILE"`
	ClusterActiveSize     int     `yaml:"cluster_active_size"     env:"ETCD_CLUSTER_ACTIVE_SIZE"`
	ClusterRemoveDelay    float64 `yaml:"cluster_remove_delay"    env:"ETCD_CLUSTER_REMOVE_DELAY"`
	ClusterSyncInterval   float64 `yaml:"cluster_sync_interval"   env:"ETCD_CLUSTER_SYNC_INTERVAL"`
	CorsOrigins           string  `yaml:"cors"                    env:"ETCD_CORS"`
	DataDir               string  `yaml:"data_dir"                env:"ETCD_DATA_DIR"`
	Discovery             string  `yaml:"discovery"               env:"ETCD_DISCOVERY"`
	GraphiteHost          string  `yaml:"graphite_host"           env:"ETCD_GRAPHITE_HOST"`
	HTTPReadTimeout       float64 `yaml:"http_read_timeout"       env:"ETCD_HTTP_READ_TIMEOUT"`
	HTTPWriteTimeout      float64 `yaml:"http_write_timeout"      env:"ETCD_HTTP_WRITE_TIMEOUT"`
	KeyFile               string  `yaml:"key_file"                env:"ETCD_KEY_FILE"`
	MaxResultBuffer       int     `yaml:"max_result_buffer"       env:"ETCD_MAX_RESULT_BUFFER"`
	MaxRetryAttempts      int     `yaml:"max_retry_attempts"      env:"ETCD_MAX_RETRY_ATTEMPTS"`
	Name                  string  `yaml:"name"                    env:"ETCD_NAME"`
	PeerAddr              string  `yaml:"peer_addr"               env:"ETCD_PEER_ADDR"`
	PeerBindAddr          string  `yaml:"peer_bind_addr"          env:"ETCD_PEER_BIND_ADDR"`
	PeerCAFile            string  `yaml:"peer_ca_file"            env:"ETCD_PEER_CA_FILE"`
	PeerCertFile          string  `yaml:"peer_cert_file"          env:"ETCD_PEER_CERT_FILE"`
	PeerElectionTimeout   int     `yaml:"peer_election_timeout"   env:"ETCD_PEER_ELECTION_TIMEOUT"`
	PeerHeartbeatInterval int     `yaml:"peer_heartbeat_interval" env:"ETCD_PEER_HEARTBEAT_INTERVAL"`
	PeerKeyFile           string  `yaml:"peer_key_file"           env:"ETCD_PEER_KEY_FILE"`
	Peers                 string  `yaml:"peers"                   env:"ETCD_PEERS"`
	PeersFile             string  `yaml:"peers_file"              env:"ETCD_PEERS_FILE"`
	RetryInterval         float64 `yaml:"retry_interval"          env:"ETCD_RETRY_INTERVAL"`
	Snapshot              bool    `yaml:"snapshot"                env:"ETCD_SNAPSHOT"`
	SnapshotCount         int     `yaml:"snapshot_count"          env:"ETCD_SNAPSHOTCOUNT"`
	StrTrace              string  `yaml:"trace"                   env:"ETCD_TRACE"`
	Verbose               bool    `yaml:"verbose"                 env:"ETCD_VERBOSE"`
	VeryVerbose           bool    `yaml:"very_verbose"            env:"ETCD_VERY_VERBOSE"`
	VeryVeryVerbose       bool    `yaml:"very_very_verbose"       env:"ETCD_VERY_VERY_VERBOSE"`
}
