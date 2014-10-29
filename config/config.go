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

package config

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/coreos/coreos-cloudinit/Godeps/_workspace/src/gopkg.in/yaml.v1"
)

// CloudConfig encapsulates the entire cloud-config configuration file and maps
// directly to YAML. Fields that cannot be set in the cloud-config (fields
// used for internal use) have the YAML tag '-' so that they aren't marshalled.
type CloudConfig struct {
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	Coreos            struct {
		Etcd   Etcd   `yaml:"etcd"`
		Fleet  Fleet  `yaml:"fleet"`
		OEM    OEM    `yaml:"oem"`
		Update Update `yaml:"update"`
		Units  []Unit `yaml:"units"`
	} `yaml:"coreos"`
	WriteFiles        []File   `yaml:"write_files"`
	Hostname          string   `yaml:"hostname"`
	Users             []User   `yaml:"users"`
	ManageEtcHosts    EtcHosts `yaml:"manage_etc_hosts"`
	NetworkConfigPath string   `yaml:"-"`
	NetworkConfig     string   `yaml:"-"`
}

func IsCloudConfig(userdata string) bool {
	header := strings.SplitN(userdata, "\n", 2)[0]

	// Explicitly trim the header so we can handle user-data from
	// non-unix operating systems. The rest of the file is parsed
	// by yaml, which correctly handles CRLF.
	header = strings.TrimSuffix(header, "\r")

	return (header == "#cloud-config")
}

// NewCloudConfig instantiates a new CloudConfig from the given contents (a
// string of YAML), returning any error encountered. It will ignore unknown
// fields but log encountering them.
func NewCloudConfig(contents string) (*CloudConfig, error) {
	var cfg CloudConfig
	ncontents, err := normalizeConfig(contents)
	if err != nil {
		return &cfg, err
	}
	if err = yaml.Unmarshal(ncontents, &cfg); err != nil {
		return &cfg, err
	}
	warnOnUnrecognizedKeys(contents, log.Printf)
	return &cfg, nil
}

func (cc CloudConfig) String() string {
	bytes, err := yaml.Marshal(cc)
	if err != nil {
		return ""
	}

	stringified := string(bytes)
	stringified = fmt.Sprintf("#cloud-config\n%s", stringified)

	return stringified
}

// IsZero returns whether or not the parameter is the zero value for its type.
// If the parameter is a struct, only the exported fields are considered.
func IsZero(c interface{}) bool {
	return isZero(reflect.ValueOf(c))
}

type ErrorValid struct {
	Value string
	Valid []string
	Field string
}

func (e ErrorValid) Error() string {
	return fmt.Sprintf("invalid value %q for option %q (valid options: %q)", e.Value, e.Field, e.Valid)
}

// AssertStructValid checks the fields in the structure and makes sure that
// they contain valid values as specified by the 'valid' flag. Empty fields are
// implicitly valid.
func AssertStructValid(c interface{}) error {
	ct := reflect.TypeOf(c)
	cv := reflect.ValueOf(c)
	for i := 0; i < ct.NumField(); i++ {
		ft := ct.Field(i)
		if !isFieldExported(ft) {
			continue
		}

		if err := AssertValid(cv.Field(i), ft.Tag.Get("valid")); err != nil {
			err.Field = ft.Name
			return err
		}
	}
	return nil
}

// AssertValid checks to make sure that the given value is in the list of
// valid values. Zero values are implicitly valid.
func AssertValid(value reflect.Value, valid string) *ErrorValid {
	if valid == "" || isZero(value) {
		return nil
	}
	vs := fmt.Sprintf("%v", value.Interface())
	valids := strings.Split(valid, ",")
	for _, valid := range valids {
		if vs == valid {
			return nil
		}
	}
	return &ErrorValid{
		Value: vs,
		Valid: valids,
	}
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Struct:
		vt := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if isFieldExported(vt.Field(i)) && !isZero(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return v.Interface() == reflect.Zero(v.Type()).Interface()
	}
}

func isFieldExported(f reflect.StructField) bool {
	return f.PkgPath == ""
}

type warner func(format string, v ...interface{})

// warnOnUnrecognizedKeys parses the contents of a cloud-config file and calls
// warn(msg, key) for every unrecognized key (i.e. those not present in CloudConfig)
func warnOnUnrecognizedKeys(contents string, warn warner) {
	// Generate a map of all understood cloud config options
	var cc map[string]interface{}
	b, _ := yaml.Marshal(&CloudConfig{})
	yaml.Unmarshal(b, &cc)

	// Now unmarshal the entire provided contents
	var c map[string]interface{}
	yaml.Unmarshal([]byte(contents), &c)

	// Check that every key in the contents exists in the cloud config
	for k, _ := range c {
		if _, ok := cc[k]; !ok {
			warn("Warning: unrecognized key %q in provided cloud config - ignoring section", k)
		}
	}

	// Check for unrecognized coreos options, if any are set
	if coreos, ok := c["coreos"]; ok {
		if set, ok := coreos.(map[interface{}]interface{}); ok {
			known := cc["coreos"].(map[interface{}]interface{})
			for k, _ := range set {
				if key, ok := k.(string); ok {
					if _, ok := known[key]; !ok {
						warn("Warning: unrecognized key %q in coreos section of provided cloud config - ignoring", key)
					}
				} else {
					warn("Warning: unrecognized key %q in coreos section of provided cloud config - ignoring", k)
				}
			}
		}
	}

	// Check for any badly-specified users, if any are set
	if users, ok := c["users"]; ok {
		var known map[string]interface{}
		b, _ := yaml.Marshal(&User{})
		yaml.Unmarshal(b, &known)

		if set, ok := users.([]interface{}); ok {
			for _, u := range set {
				if user, ok := u.(map[interface{}]interface{}); ok {
					for k, _ := range user {
						if key, ok := k.(string); ok {
							if _, ok := known[key]; !ok {
								warn("Warning: unrecognized key %q in user section of cloud config - ignoring", key)
							}
						} else {
							warn("Warning: unrecognized key %q in user section of cloud config - ignoring", k)
						}
					}
				}
			}
		}
	}

	// Check for any badly-specified files, if any are set
	if files, ok := c["write_files"]; ok {
		var known map[string]interface{}
		b, _ := yaml.Marshal(&File{})
		yaml.Unmarshal(b, &known)

		if set, ok := files.([]interface{}); ok {
			for _, f := range set {
				if file, ok := f.(map[interface{}]interface{}); ok {
					for k, _ := range file {
						if key, ok := k.(string); ok {
							if _, ok := known[key]; !ok {
								warn("Warning: unrecognized key %q in file section of cloud config - ignoring", key)
							}
						} else {
							warn("Warning: unrecognized key %q in file section of cloud config - ignoring", k)
						}
					}
				}
			}
		}
	}
}

func normalizeConfig(config string) ([]byte, error) {
	var cfg map[interface{}]interface{}
	if err := yaml.Unmarshal([]byte(config), &cfg); err != nil {
		return nil, err
	}
	return yaml.Marshal(normalizeKeys(cfg))
}

func normalizeKeys(m map[interface{}]interface{}) map[interface{}]interface{} {
	for k, v := range m {
		if m, ok := m[k].(map[interface{}]interface{}); ok {
			normalizeKeys(m)
		}

		if s, ok := m[k].([]interface{}); ok {
			for _, e := range s {
				if m, ok := e.(map[interface{}]interface{}); ok {
					normalizeKeys(m)
				}
			}
		}

		delete(m, k)
		m[strings.Replace(fmt.Sprint(k), "-", "_", -1)] = v
	}
	return m
}
