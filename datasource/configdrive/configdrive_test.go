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

package configdrive

import (
	"bytes"
	"testing"

	"github.com/coreos/coreos-cloudinit/datasource/test"
)

func TestNewDatasource(t *testing.T) {
	tests := []struct {
		root string

		expectRoot string
	}{
		{
			root:       "",
			expectRoot: "",
		},
		{
			root:       "/media/configdrive",
			expectRoot: "/media/configdrive",
		},
	}

	for i, tt := range tests {
		service := NewDatasource(tt.root)
		if service.Root != tt.expectRoot {
			t.Errorf("bad root (test %d): want %q, got %q", i, tt.expectRoot, service.Root)
		}
	}
}

func TestConfigRoot(t *testing.T) {
	tests := []struct {
		root string

		configRoot string
	}{
		{
			"/",
			"/",
		},
		{
			"/media/configdrive",
			"/media/configdrive",
		},
	}

	for i, tt := range tests {
		cd := ConfigDrive{tt.root, nil}
		if configRoot := cd.ConfigRoot(); configRoot != tt.configRoot {
			t.Errorf("bad config root (test %d): want %q, got %q", i, tt.configRoot, configRoot)
		}
	}
}

func TestTryReadFile(t *testing.T) {
	tests := []struct {
		filename string
		files    test.MockFilesystem

		content []byte
	}{
		{
			filename: "/test",
			files:    test.MockFilesystem{"/test": "my file"},
			content:  []byte("my file"),
		},
		{
			filename: "/test",
		},
	}

	for i, tt := range tests {
		service := ConfigDrive{"", tt.files.ReadFile}
		content, err := service.TryReadFile(tt.filename)
		if err != nil {
			t.Errorf("bad error (test %d): want %v, got %v", i, nil, err)
		}
		if !bytes.Equal(content, tt.content) {
			t.Errorf("bad root (test %d): want %q, got %q", i, tt.content, content)
		}
	}
}
