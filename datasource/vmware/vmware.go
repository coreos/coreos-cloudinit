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

package vmware

import (
	"fmt"
	"log"
	"net"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/datasource"

	"github.com/coreos/coreos-cloudinit/Godeps/_workspace/src/github.com/sigma/vmw-guestinfo/rpcvmx"
	"github.com/coreos/coreos-cloudinit/Godeps/_workspace/src/github.com/sigma/vmw-guestinfo/vmcheck"
)

type vmware struct {
	readConfig func(key string) (string, error)
}

func NewDatasource() *vmware {
	return &vmware{readConfig}
}

func (v vmware) IsAvailable() bool {
	return vmcheck.IsVirtualWorld()
}

func (v vmware) AvailabilityChanges() bool {
	return false
}

func (v vmware) ConfigRoot() string {
	return "/"
}

func (v vmware) FetchMetadata() (metadata datasource.Metadata, err error) {
	metadata.Hostname, _ = v.readConfig("hostname")

	netconf := map[string]string{}
	saveConfig := func(key string, args ...interface{}) string {
		key = fmt.Sprintf(key, args...)
		val, _ := v.readConfig(key)
		if val != "" {
			netconf[key] = val
		}
		return val
	}

	for i := 0; ; i++ {
		if nameserver := saveConfig("dns.server.%d", i); nameserver == "" {
			break
		}
	}

	found := true
	for i := 0; found; i++ {
		found = false

		found = (saveConfig("interface.%d.name", i) != "") || found
		found = (saveConfig("interface.%d.mac", i) != "") || found
		found = (saveConfig("interface.%d.dhcp", i) != "") || found

		role, _ := v.readConfig(fmt.Sprintf("interface.%d.role", i))
		for a := 0; ; a++ {
			address := saveConfig("interface.%d.ip.%d.address", i, a)
			if address == "" {
				break
			} else {
				found = true
			}

			ip, _, err := net.ParseCIDR(address)
			if err != nil {
				return metadata, err
			}

			switch role {
			case "public":
				if ip.To4() != nil {
					metadata.PublicIPv4 = ip
				} else {
					metadata.PublicIPv6 = ip
				}
			case "private":
				if ip.To4() != nil {
					metadata.PrivateIPv4 = ip
				} else {
					metadata.PrivateIPv6 = ip
				}
			case "":
			default:
				return metadata, fmt.Errorf("unrecognized role: %q", role)
			}
		}

		for r := 0; ; r++ {
			gateway := saveConfig("interface.%d.route.%d.gateway", i, r)
			destination := saveConfig("interface.%d.route.%d.destination", i, r)

			if gateway == "" && destination == "" {
				break
			} else {
				found = true
			}
		}
	}
	metadata.NetworkConfig = netconf

	return
}

func (v vmware) FetchUserdata() ([]byte, error) {
	encoding, err := v.readConfig("coreos.config.data.encoding")
	if err != nil {
		return nil, err
	}

	data, err := v.readConfig("coreos.config.data")
	if err != nil {
		return nil, err
	}

	if encoding != "" {
		return config.DecodeContent(data, encoding)
	}
	return []byte(data), nil
}

func (v vmware) Type() string {
	return "vmware"
}

func readConfig(key string) (string, error) {
	data, err := rpcvmx.NewConfig().GetString(key, "")
	if err == nil {
		log.Printf("Read from %q: %q\n", key, data)
	} else {
		log.Printf("Failed to read from %q: %v\n", key, err)
	}
	return data, err
}
