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

package waagent

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
)

type waagent struct {
	root     string
	readFile func(filename string) ([]byte, error)
}

func NewDatasource(root string) *waagent {
	return &waagent{root, ioutil.ReadFile}
}

func (a *waagent) IsAvailable() bool {
	_, err := os.Stat(path.Join(a.root, "provisioned"))
	return !os.IsNotExist(err)
}

func (a *waagent) AvailabilityChanges() bool {
	return true
}

func (a *waagent) ConfigRoot() string {
	return a.root
}

func (a *waagent) FetchMetadata() ([]byte, error) {
	metadataBytes, err := a.tryReadFile(path.Join(a.root, "SharedConfig.xml"))
	if err != nil {
		return nil, err
	}
	if len(metadataBytes) == 0 {
		return metadataBytes, nil
	}

	type Instance struct {
		Id             string `xml:"id,attr"`
		Address        string `xml:"address,attr"`
		InputEndpoints struct {
			Endpoints []struct {
				LoadBalancedPublicAddress string `xml:"loadBalancedPublicAddress,attr"`
			} `xml:"Endpoint"`
		}
	}

	type SharedConfig struct {
		Incarnation struct {
			Instance string `xml:"instance,attr"`
		}
		Instances struct {
			Instances []Instance `xml:"Instance"`
		}
	}

	var metadata SharedConfig
	if err := xml.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, err
	}

	var instance Instance
	for _, i := range metadata.Instances.Instances {
		if i.Id == metadata.Incarnation.Instance {
			instance = i
			break
		}
	}

	attrs := map[string]string{
		"local-ipv4": instance.Address,
	}
	for _, e := range instance.InputEndpoints.Endpoints {
		host, _, err := net.SplitHostPort(e.LoadBalancedPublicAddress)
		if err == nil {
			attrs["public-ipv4"] = host
			break
		}
	}
	return json.Marshal(attrs)
}

func (a *waagent) FetchUserdata() ([]byte, error) {
	return a.tryReadFile(path.Join(a.root, "CustomData"))
}

func (a *waagent) FetchNetworkConfig(filename string) ([]byte, error) {
	return nil, nil
}

func (a *waagent) Type() string {
	return "waagent"
}

func (a *waagent) tryReadFile(filename string) ([]byte, error) {
	fmt.Printf("Attempting to read from %q\n", filename)
	data, err := a.readFile(filename)
	if os.IsNotExist(err) {
		err = nil
	}
	return data, err
}
