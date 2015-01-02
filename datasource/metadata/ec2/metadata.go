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

package ec2

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coreos/coreos-cloudinit/datasource/metadata"
	"github.com/coreos/coreos-cloudinit/pkg"
)

const (
	DefaultAddress = "http://169.254.169.254/"
	apiVersion     = "2009-04-04/"
	userdataPath   = apiVersion + "user-data"
	metadataPath   = apiVersion + "meta-data"
)

type metadataService struct {
	metadata.MetadataService
}

func NewDatasource(root string) *metadataService {
	return &metadataService{metadata.NewDatasource(root, apiVersion, userdataPath, metadataPath)}
}

func (ms metadataService) FetchMetadata() ([]byte, error) {
	attrs := make(map[string]interface{})
	if keynames, err := ms.fetchAttributes(fmt.Sprintf("%s/public-keys", ms.MetadataUrl())); err == nil {
		keyIDs := make(map[string]string)
		for _, keyname := range keynames {
			tokens := strings.SplitN(keyname, "=", 2)
			if len(tokens) != 2 {
				return nil, fmt.Errorf("malformed public key: %q", keyname)
			}
			keyIDs[tokens[1]] = tokens[0]
		}

		keys := make(map[string]string)
		for name, id := range keyIDs {
			sshkey, err := ms.fetchAttribute(fmt.Sprintf("%s/public-keys/%s/openssh-key", ms.MetadataUrl(), id))
			if err != nil {
				return nil, err
			}
			keys[name] = sshkey
			fmt.Printf("Found SSH key for %q\n", name)
		}
		attrs["public_keys"] = keys
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if hostname, err := ms.fetchAttribute(fmt.Sprintf("%s/hostname", ms.MetadataUrl())); err == nil {
		attrs["hostname"] = strings.Split(hostname, " ")[0]
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if localAddr, err := ms.fetchAttribute(fmt.Sprintf("%s/local-ipv4", ms.MetadataUrl())); err == nil {
		attrs["local-ipv4"] = localAddr
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if publicAddr, err := ms.fetchAttribute(fmt.Sprintf("%s/public-ipv4", ms.MetadataUrl())); err == nil {
		attrs["public-ipv4"] = publicAddr
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if content_path, err := ms.fetchAttribute(fmt.Sprintf("%s/network_config/content_path", ms.MetadataUrl())); err == nil {
		attrs["network_config"] = map[string]string{
			"content_path": content_path,
		}
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	return json.Marshal(attrs)
}

func (ms metadataService) Type() string {
	return "ec2-metadata-service"
}

func (ms metadataService) fetchAttributes(url string) ([]string, error) {
	resp, err := ms.FetchData(url)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(resp))
	data := make([]string, 0)
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}
	return data, scanner.Err()
}

func (ms metadataService) fetchAttribute(url string) (string, error) {
	if attrs, err := ms.fetchAttributes(url); err == nil && len(attrs) > 0 {
		return attrs[0], nil
	} else {
		return "", err
	}
}
