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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/coreos/coreos-cloudinit/datasource"
)

type ConfigDrive struct {
	Root     string
	ReadFile func(filename string) ([]byte, error)
}

func NewDatasource(root string) ConfigDrive {
	return ConfigDrive{root, ioutil.ReadFile}
}

func (cd *ConfigDrive) IsAvailable() bool {
	_, err := os.Stat(cd.Root)
	return !os.IsNotExist(err)
}

func (cd *ConfigDrive) AvailabilityChanges() bool {
	return true
}

func (cd *ConfigDrive) ConfigRoot() string {
	return cd.Root
}

func (cd *ConfigDrive) FetchMetadata() (datasource.Metadata, error) {
	return datasource.Metadata{}, nil
}

func (cd *ConfigDrive) FetchUserdata() ([]byte, error) {
	return nil, nil
}

func (cd *ConfigDrive) Type() string {
	return "cloud-drive"
}

func (cd *ConfigDrive) TryReadFile(filename string) ([]byte, error) {
	fmt.Printf("Attempting to read from %q\n", filename)
	data, err := cd.ReadFile(filename)
	if os.IsNotExist(err) {
		err = nil
	}
	return data, err
}
