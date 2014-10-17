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
	"bytes"
	"fmt"
	"testing"

	"github.com/coreos/coreos-cloudinit/datasource/metadata"
	"github.com/coreos/coreos-cloudinit/datasource/metadata/test"
	"github.com/coreos/coreos-cloudinit/pkg"
)

func TestType(t *testing.T) {
	want := "digitalocean-metadata-service"
	if kind := (metadataService{}).Type(); kind != want {
		t.Fatalf("bad type: want %q, got %q", want, kind)
	}
}

func TestFetchMetadata(t *testing.T) {
	for _, tt := range []struct {
		root         string
		metadataPath string
		resources    map[string]string
		expect       []byte
		clientErr    error
		expectErr    error
	}{
		{
			root:         "/",
			metadataPath: "v1.json",
			resources: map[string]string{
				"/v1.json": "bad",
			},
			expectErr: fmt.Errorf("invalid character 'b' looking for beginning of value"),
		},
		{
			root:         "/",
			metadataPath: "v1.json",
			resources: map[string]string{
				"/v1.json": `{
  "droplet_id": 1,
  "user_data": "hello",
  "vendor_data": "hello",
  "public_keys": [
    "publickey1",
    "publickey2"
  ],
  "region": "nyc2",
  "interfaces": {
    "public": [
      {
        "ipv4": {
          "ip_address": "192.168.1.2",
          "netmask": "255.255.255.0",
          "gateway": "192.168.1.1"
        },
        "ipv6": {
          "ip_address": "fe00::",
          "cidr": 126,
          "gateway": "fe00::"
        },
        "mac": "ab:cd:ef:gh:ij",
        "type": "public"
      }
    ]
  }
}`,
			},
			expect: []byte(`{"hostname":"","public-ipv4":"192.168.1.2","public-ipv6":"fe00::","public_keys":{"0":"publickey1","1":"publickey2"}}`),
		},
		{
			clientErr: pkg.ErrTimeout{fmt.Errorf("test error")},
			expectErr: pkg.ErrTimeout{fmt.Errorf("test error")},
		},
	} {
		service := &metadataService{
			MetadataService: metadata.MetadataService{
				Root:         tt.root,
				Client:       &test.HttpClient{tt.resources, tt.clientErr},
				MetadataPath: tt.metadataPath,
			},
		}
		metadata, err := service.FetchMetadata()
		if Error(err) != Error(tt.expectErr) {
			t.Fatalf("bad error (%q): want %q, got %q", tt.resources, tt.expectErr, err)
		}
		if !bytes.Equal(metadata, tt.expect) {
			t.Fatalf("bad fetch (%q): want %q, got %q", tt.resources, tt.expect, metadata)
		}
	}
}

func Error(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
