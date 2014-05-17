package initialize

import (
	"errors"
	"fmt"

	"github.com/coreos/coreos-cloudinit/system"
)

type EtcdEnvironment map[string]string

func (ee EtcdEnvironment) String() (out string) {
	norm := normalizeSvcEnv(ee)

	if val, ok := norm["DISCOVERY_URL"]; ok {
		delete(norm, "DISCOVERY_URL")
		if _, ok := norm["DISCOVERY"]; !ok {
			norm["DISCOVERY"] = val
		}
	}

	out += "[Service]\n"

	for key, val := range norm {
		out += fmt.Sprintf("Environment=\"ETCD_%s=%s\"\n", key, val)
	}

	return
}

// Unit creates a Unit file drop-in for etcd, using any configured
// options and adding a default MachineID if unset.
func (ee EtcdEnvironment) Unit(root string) (*system.Unit, error) {
	if ee == nil {
		return nil, nil
	}

	if _, ok := ee["name"]; !ok {
		if machineID := system.MachineID(root); machineID != "" {
			ee["name"] = machineID
		} else if hostname, err := system.Hostname(); err == nil {
			ee["name"] = hostname
		} else {
			return nil, errors.New("Unable to determine default etcd name")
		}
	}

	return &system.Unit{
		Name:    "etcd.service",
		Runtime: true,
		DropIn:  true,
		Content: ee.String(),
	}, nil
}
