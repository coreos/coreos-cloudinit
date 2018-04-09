// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

type Flannel struct {
	EtcdEndpoints string `yaml:"etcd_endpoints,omitempty" env:"FLANNELD_ETCD_ENDPOINTS"`
	EtcdCAFile    string `yaml:"etcd_cafile,omitempty"    env:"FLANNELD_ETCD_CAFILE"`
	EtcdCertFile  string `yaml:"etcd_certfile,omitempty"  env:"FLANNELD_ETCD_CERTFILE"`
	EtcdKeyFile   string `yaml:"etcd_keyfile,omitempty"   env:"FLANNELD_ETCD_KEYFILE"`
	EtcdPrefix    string `yaml:"etcd_prefix,omitempty"    env:"FLANNELD_ETCD_PREFIX"`
	EtcdUsername  string `yaml:"etcd_username,omitempty"  env:"FLANNELD_ETCD_USERNAME"`
	EtcdPassword  string `yaml:"etcd_password,omitempty"  env:"FLANNELD_ETCD_PASSWORD"`
	IPMasq        string `yaml:"ip_masq,omitempty"        env:"FLANNELD_IP_MASQ"`
	SubnetFile    string `yaml:"subnet_file,omitempty"    env:"FLANNELD_SUBNET_FILE"`
	Iface         string `yaml:"interface,omitempty"      env:"FLANNELD_IFACE"`
	PublicIP      string `yaml:"public_ip,omitempty"      env:"FLANNELD_PUBLIC_IP"`
}
