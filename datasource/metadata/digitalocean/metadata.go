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

package digitalocean

import (
	"encoding/json"
	"strconv"

	"github.com/coreos/coreos-cloudinit/datasource/metadata"
)

const (
	DefaultAddress = "http://169.254.169.254/"
	apiVersion     = "metadata/v1"
	userdataUrl    = apiVersion + "/user-data"
	metadataPath   = apiVersion + ".json"
)

type Address struct {
	IPAddress string `json:"ip_address"`
	Netmask   string `json:"netmask"`
	Cidr      int    `json:"cidr"`
	Gateway   string `json:"gateway"`
}

type Interface struct {
	IPv4 *Address `json:"ipv4"`
	IPv6 *Address `json:"ipv6"`
	MAC  string   `json:"mac"`
	Type string   `json:"type"`
}

type Interfaces struct {
	Public  []Interface `json:"public"`
	Private []Interface `json:"private"`
}

type DNS struct {
	Nameservers []string `json:"nameservers"`
}

type Metadata struct {
	Hostname   string     `json:"hostname"`
	Interfaces Interfaces `json:"interfaces"`
	PublicKeys []string   `json:"public_keys"`
	DNS        DNS        `json:"dns"`
}

type metadataService struct {
	interfaces Interfaces
	dns        DNS
	metadata.MetadataService
}

func NewDatasource(root string) *metadataService {
	return &metadataService{MetadataService: metadata.NewDatasource(root, apiVersion, userdataUrl, metadataPath)}
}

func (ms *metadataService) FetchMetadata() ([]byte, error) {
	data, err := ms.FetchData(ms.MetadataUrl())
	if err != nil || len(data) == 0 {
		return []byte{}, err
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return []byte{}, err
	}

	ms.interfaces = metadata.Interfaces
	ms.dns = metadata.DNS

	attrs := make(map[string]interface{})
	if len(metadata.Interfaces.Public) > 0 {
		if metadata.Interfaces.Public[0].IPv4 != nil {
			attrs["public-ipv4"] = metadata.Interfaces.Public[0].IPv4.IPAddress
		}
		if metadata.Interfaces.Public[0].IPv6 != nil {
			attrs["public-ipv6"] = metadata.Interfaces.Public[0].IPv6.IPAddress
		}
	}
	if len(metadata.Interfaces.Private) > 0 {
		if metadata.Interfaces.Private[0].IPv4 != nil {
			attrs["local-ipv4"] = metadata.Interfaces.Private[0].IPv4.IPAddress
		}
		if metadata.Interfaces.Private[0].IPv6 != nil {
			attrs["local-ipv6"] = metadata.Interfaces.Private[0].IPv6.IPAddress
		}
	}
	attrs["hostname"] = metadata.Hostname
	keys := make(map[string]string)
	for i, key := range metadata.PublicKeys {
		keys[strconv.Itoa(i)] = key
	}
	attrs["public_keys"] = keys

	return json.Marshal(attrs)
}

func (ms metadataService) FetchNetworkConfig(filename string) ([]byte, error) {
	return json.Marshal(Metadata{
		Interfaces: ms.interfaces,
		DNS:        ms.dns,
	})
}

func (ms metadataService) Type() string {
	return "digitalocean-metadata-service"
}
