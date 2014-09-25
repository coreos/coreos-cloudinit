package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/coreos/coreos-cloudinit/third_party/gopkg.in/yaml.v1"
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

// NewCloudConfig instantiates a new CloudConfig from the given contents (a
// string of YAML), returning any error encountered. It will ignore unknown
// fields but log encountering them.
func NewCloudConfig(contents string) (*CloudConfig, error) {
	var cfg CloudConfig
	err := yaml.Unmarshal([]byte(contents), &cfg)
	if err != nil {
		return &cfg, err
	}
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

// AssertValid checks the fields in the structure and makes sure that they
// contain valid values as specified by the 'valid' flag. Empty fields are
// implicitly valid.
func AssertValid(c interface{}) error {
	ct := reflect.TypeOf(c)
	cv := reflect.ValueOf(c)
	for i := 0; i < ct.NumField(); i++ {
		ft := ct.Field(i)
		if !isFieldExported(ft) {
			continue
		}

		valid := ft.Tag.Get("valid")
		val := cv.Field(i)
		if !isValid(val, valid) {
			return fmt.Errorf("invalid value \"%v\" for option %q (valid options: %q)", val.Interface(), ft.Name, valid)
		}
	}
	return nil
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

func isValid(v reflect.Value, valid string) bool {
	if valid == "" || isZero(v) {
		return true
	}
	vs := fmt.Sprintf("%v", v.Interface())
	for _, valid := range strings.Split(valid, ",") {
		if vs == valid {
			return true
		}
	}
	return false
}
