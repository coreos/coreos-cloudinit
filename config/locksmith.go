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

type Locksmith struct {
	Endpoint           string `yaml:"endpoint,omitempty"      env:"LOCKSMITHD_ENDPOINT"`
	EtcdCAFile         string `yaml:"etcd_cafile,omitempty"   env:"LOCKSMITHD_ETCD_CAFILE"`
	EtcdCertFile       string `yaml:"etcd_certfile,omitempty" env:"LOCKSMITHD_ETCD_CERTFILE"`
	EtcdKeyFile        string `yaml:"etcd_keyfile,omitempty"  env:"LOCKSMITHD_ETCD_KEYFILE"`
	EtcdUsername       string `yaml:"etcd_username,omitempty" env:"LOCKSMITHD_ETCD_USERNAME"`
	EtcdPassword       string `yaml:"etcd_password,omitempty" env:"LOCKSMITHD_ETCD_PASSWORD"`
	Group              string `yaml:"group,omitempty"         env:"LOCKSMITHD_GROUP"`
	RebootWindowStart  string `yaml:"window_start,omitempty"  env:"REBOOT_WINDOW_START"  valid:"^((?i:sun|mon|tue|wed|thu|fri|sat|sun) )?0*([0-9]|1[0-9]|2[0-3]):0*([0-9]|[1-5][0-9])$"`
	RebootWindowLength string `yaml:"window_length,omitempty" env:"REBOOT_WINDOW_LENGTH" valid:"^[-+]?([0-9]*(\\.[0-9]*)?[a-z]+)+$"`
}
