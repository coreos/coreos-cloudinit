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

package metadata

import (
	"net/http"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/pkg"
)

type MetadataService struct {
	Root         string
	Client       pkg.Getter
	ApiVersion   string
	UserdataPath string
	MetadataPath string
}

func NewDatasource(root, apiVersion, userdataPath, metadataPath string, header http.Header) MetadataService {
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}
	return MetadataService{root, pkg.NewHttpClientHeader(header), apiVersion, userdataPath, metadataPath}
}

func (ms MetadataService) IsAvailable() bool {
	_, err := ms.Client.Get(ms.Root + ms.ApiVersion)
	return (err == nil)
}

func (ms MetadataService) AvailabilityChanges() bool {
	return true
}

func (ms MetadataService) ConfigRoot() string {
	return ms.Root
}

func (ms MetadataService) FetchUserdata() ([]byte, error) {
	if data, err := ms.FetchData(ms.UserdataUrl()); err != nil {
		return data, err
	} else if t_data, t_err := config.DecodeGzipContent(string(data[:])); t_err != nil {
		return data, err
	} else {
		return t_data, t_err
	}
}

func (ms MetadataService) FetchData(url string) ([]byte, error) {
	if data, err := ms.Client.GetRetry(url); err == nil {
		return data, err
	} else if _, ok := err.(pkg.ErrNotFound); ok {
		return []byte{}, nil
	} else {
		return data, err
	}
}

func (ms MetadataService) MetadataUrl() string {
	return (ms.Root + ms.MetadataPath)
}

func (ms MetadataService) UserdataUrl() string {
	return (ms.Root + ms.UserdataPath)
}
