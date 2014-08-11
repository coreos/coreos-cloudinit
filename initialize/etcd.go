package initialize

import (
	"errors"
	"fmt"
	"sort"

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

	var sorted sort.StringSlice
	for k, _ := range norm {
		sorted = append(sorted, k)
	}
	sorted.Sort()

	out += "[Service]\n"

	for _, key := range sorted {
		val := norm[key]
		out += fmt.Sprintf("Environment=\"ETCD_%s=%s\"\n", key, val)
	}

	return
}

// Units creates a Unit file drop-in for etcd, using any configured
// options and adding a default MachineID if unset.
func (ee EtcdEnvironment) Units(root string) ([]system.Unit, error) {
	if len(ee) < 1 {
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

	etcd := system.Unit{
		Name:    "etcd.service",
		Runtime: true,
		DropIn:  true,
		Content: ee.String(),
	}
	return []system.Unit{etcd}, nil
}
