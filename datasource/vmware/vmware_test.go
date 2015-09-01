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
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/datasource"
)

type MockHypervisor map[string]string

func (h MockHypervisor) ReadConfig(key string) (string, error) {
	return h[key], nil
}

func TestFetchMetadata(t *testing.T) {
	tests := []struct {
		variables MockHypervisor

		metadata datasource.Metadata
		err      error
	}{
		{
			variables: map[string]string{
				"interface.0.mac":  "test mac",
				"interface.0.dhcp": "yes",
			},
			metadata: datasource.Metadata{
				NetworkConfig: map[string]string{
					"interface.0.mac":  "test mac",
					"interface.0.dhcp": "yes",
				},
			},
		},
		{
			variables: map[string]string{
				"interface.0.name": "test name",
				"interface.0.dhcp": "yes",
			},
			metadata: datasource.Metadata{
				NetworkConfig: map[string]string{
					"interface.0.name": "test name",
					"interface.0.dhcp": "yes",
				},
			},
		},
		{
			variables: map[string]string{
				"hostname":                        "test host",
				"interface.0.mac":                 "test mac",
				"interface.0.role":                "private",
				"interface.0.ip.0.address":        "fe00::100/64",
				"interface.0.route.0.gateway":     "fe00::1",
				"interface.0.route.0.destination": "::",
			},
			metadata: datasource.Metadata{
				Hostname:    "test host",
				PrivateIPv6: net.ParseIP("fe00::100"),
				NetworkConfig: map[string]string{
					"interface.0.mac":                 "test mac",
					"interface.0.ip.0.address":        "fe00::100/64",
					"interface.0.route.0.gateway":     "fe00::1",
					"interface.0.route.0.destination": "::",
				},
			},
		},
		{
			variables: map[string]string{
				"hostname":                        "test host",
				"interface.0.name":                "test name",
				"interface.0.role":                "public",
				"interface.0.ip.0.address":        "10.0.0.100/24",
				"interface.0.ip.1.address":        "10.0.0.101/24",
				"interface.0.route.0.gateway":     "10.0.0.1",
				"interface.0.route.0.destination": "0.0.0.0",
				"interface.1.mac":                 "test mac",
				"interface.1.role":                "private",
				"interface.1.route.0.gateway":     "10.0.0.2",
				"interface.1.route.0.destination": "0.0.0.0",
				"interface.1.ip.0.address":        "10.0.0.102/24",
			},
			metadata: datasource.Metadata{
				Hostname:    "test host",
				PublicIPv4:  net.ParseIP("10.0.0.101"),
				PrivateIPv4: net.ParseIP("10.0.0.102"),
				NetworkConfig: map[string]string{
					"interface.0.name":                "test name",
					"interface.0.ip.0.address":        "10.0.0.100/24",
					"interface.0.ip.1.address":        "10.0.0.101/24",
					"interface.0.route.0.gateway":     "10.0.0.1",
					"interface.0.route.0.destination": "0.0.0.0",
					"interface.1.mac":                 "test mac",
					"interface.1.route.0.gateway":     "10.0.0.2",
					"interface.1.route.0.destination": "0.0.0.0",
					"interface.1.ip.0.address":        "10.0.0.102/24",
				},
			},
		},
	}

	for i, tt := range tests {
		v := vmware{tt.variables.ReadConfig}
		metadata, err := v.FetchMetadata()
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (#%d): want %v, got %v", i, tt.err, err)
		}
		if !reflect.DeepEqual(tt.metadata, metadata) {
			t.Errorf("bad metadata (#%d): want %#v, got %#v", i, tt.metadata, metadata)
		}
	}
}

func TestFetchUserdata(t *testing.T) {
	tests := []struct {
		variables MockHypervisor

		userdata string
		err      error
	}{
		{},
		{
			variables: map[string]string{"coreos.config.data": "test config"},
			userdata:  "test config",
		},
		{
			variables: map[string]string{
				"coreos.config.data.encoding": "",
				"coreos.config.data":          "test config",
			},
			userdata: "test config",
		},
		{
			variables: map[string]string{
				"coreos.config.data.encoding": "base64",
				"coreos.config.data":          "dGVzdCBjb25maWc=",
			},
			userdata: "test config",
		},
		{
			variables: map[string]string{
				"coreos.config.data.encoding": "gzip+base64",
				"coreos.config.data":          "H4sIABaoWlUAAytJLS5RSM7PS8tMBwCQiHNZCwAAAA==",
			},
			userdata: "test config",
		},
		{
			variables: map[string]string{
				"coreos.config.data.encoding": "test encoding",
			},
			err: errors.New(`Unsupported encoding "test encoding"`),
		},
	}

	for i, tt := range tests {
		v := vmware{tt.variables.ReadConfig}
		userdata, err := v.FetchUserdata()
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (#%d): want %v, got %v", i, nil, err)
		}
		if tt.userdata != string(userdata) {
			t.Errorf("bad userdata (#%d): want %q, got %q", i, tt.userdata, userdata)
		}
	}
}

func TestFetchUserdataError(t *testing.T) {
	testErr := errors.New("test error")
	_, err := vmware{func(_ string) (string, error) { return "", testErr }}.FetchUserdata()

	if testErr != err {
		t.Errorf("bad error: want %v, got %v", testErr, err)
	}
}
