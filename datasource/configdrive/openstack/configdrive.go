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

package openstack

import (
	"encoding/json"
	"path"

	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/datasource/configdrive"
)

const (
	apiVersion = "latest"
)

type configDrive struct {
	configdrive.ConfigDrive
}

func NewDatasource(root string) *configDrive {
	return &configDrive{configdrive.NewDatasource(path.Join(root, "openstack"))}
}

func (cd *configDrive) FetchMetadata() (metadata datasource.Metadata, err error) {
	var data []byte
	var m struct {
		SSHAuthorizedKeyMap map[string]string `json:"public_keys"`
		Hostname            string            `json:"hostname"`
		NetworkConfig       struct {
			ContentPath string `json:"content_path"`
		} `json:"network_config"`
	}

	if data, err = cd.TryReadFile(path.Join(cd.Root, apiVersion, "meta_data.json")); err != nil || len(data) == 0 {
		return
	}
	if err = json.Unmarshal([]byte(data), &m); err != nil {
		return
	}

	metadata.SSHPublicKeys = m.SSHAuthorizedKeyMap
	metadata.Hostname = m.Hostname
	if m.NetworkConfig.ContentPath != "" {
		metadata.NetworkConfig, err = cd.TryReadFile(path.Join(cd.Root, m.NetworkConfig.ContentPath))
	}

	return
}

func (cd *configDrive) FetchUserdata() ([]byte, error) {
	return cd.TryReadFile(path.Join(cd.Root, apiVersion, "user_data"))
}
