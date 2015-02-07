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
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/datasource/configdrive"
	"github.com/coreos/coreos-cloudinit/datasource/test"
)

func TestNewDatasource(t *testing.T) {
	tests := []struct {
		root       string
		configRoot string
	}{
		{
			"/",
			"/openstack",
		},
		{
			"/media/configdrive",
			"/media/configdrive/openstack",
		},
	}

	for i, tt := range tests {
		cd := NewDatasource(tt.root)
		if configRoot := cd.ConfigRoot(); configRoot != tt.configRoot {
			t.Fatalf("bad config root (test %d): want %q, got %q", i, tt.configRoot, configRoot)
		}
	}
}

func TestFetchMetadata(t *testing.T) {
	tests := []struct {
		root  string
		files test.MockFilesystem

		metadata datasource.Metadata
	}{
		{},
		{
			root:  "/",
			files: test.MockFilesystem{"/latest/meta_data.json": `{"ignore": "me"}`},
		},
		{
			root:     "/",
			files:    test.MockFilesystem{"/latest/meta_data.json": `{"hostname": "host"}`},
			metadata: datasource.Metadata{Hostname: "host"},
		},
		{
			root: "/media/configdrive/openstack",
			files: test.MockFilesystem{
				"/media/configdrive/openstack/latest/meta_data.json": `{"hostname": "host", "network_config": {"content_path": "config_file.json"}, "public_keys":{"1": "key1", "2": "key2"}}`,
				"/media/configdrive/openstack/config_file.json":      "make it work",
			},
			metadata: datasource.Metadata{
				Hostname:      "host",
				NetworkConfig: []byte("make it work"),
				SSHPublicKeys: map[string]string{
					"1": "key1",
					"2": "key2",
				},
			},
		},
	}

	for i, tt := range tests {
		cd := configDrive{configdrive.ConfigDrive{Root: tt.root, ReadFile: tt.files.ReadFile}}
		metadata, err := cd.FetchMetadata()
		if err != nil {
			t.Fatalf("bad error (test %d): want %v, got %q", i, nil, err)
		}
		if !reflect.DeepEqual(tt.metadata, metadata) {
			t.Fatalf("bad metadata (test %d): want %#v, got %#v", i, tt.metadata, metadata)
		}
	}
}

func TestFetchUserdata(t *testing.T) {
	tests := []struct {
		root  string
		files test.MockFilesystem

		userdata string
	}{
		{
			"/",
			test.MockFilesystem{},
			"",
		},
		{
			"/",
			test.MockFilesystem{"/latest/user_data": "userdata"},
			"userdata",
		},
		{
			"/media/configdrive/openstack",
			test.MockFilesystem{"/media/configdrive/openstack/latest/user_data": "userdata"},
			"userdata",
		},
	}

	for i, tt := range tests {
		cd := configDrive{configdrive.ConfigDrive{Root: tt.root, ReadFile: tt.files.ReadFile}}
		userdata, err := cd.FetchUserdata()
		if err != nil {
			t.Fatalf("bad error (test %d): want %v, got %q", i, nil, err)
		}
		if tt.userdata != string(userdata) {
			t.Fatalf("bad userdata (test %d): want %q, got %q", i, tt.userdata, userdata)
		}
	}
}
