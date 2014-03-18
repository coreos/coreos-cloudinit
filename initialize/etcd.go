package initialize

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/coreos/coreos-cloudinit/system"
)

type EtcdEnvironment map[string]string

func (ec EtcdEnvironment) normalized() map[string]string {
	out := make(map[string]string, len(ec))
	for key, val := range ec {
		key = strings.ToUpper(key)
		key = strings.Replace(key, "-", "_", -1)
		out[key] = val
	}
	return out
}

func (ec EtcdEnvironment) String() (out string) {
	norm := ec.normalized()

	if val, ok := norm["DISCOVERY_URL"]; ok {
		delete(norm, "DISCOVERY_URL")
		if _, ok := norm["DISCOVERY"]; !ok {
			norm["DISCOVERY"] = val
		}
	}

	public := os.Getenv("COREOS_PUBLIC_IPV4")
	private := os.Getenv("COREOS_PRIVATE_IPV4")

	out += "[Service]\n"

	for key, val := range norm {
		if public != "" {
			val = strings.Replace(val, "$public_ipv4", public, -1)
		}

		if private != "" {
			val = strings.Replace(val, "$private_ipv4", private, -1)
		}

		out += fmt.Sprintf("Environment=\"ETCD_%s=%s\"\n", key, val)
	}

	return
}

// Write an EtcdEnvironment to the appropriate path on disk for etcd.service
func WriteEtcdEnvironment(env EtcdEnvironment, root string) error {
	if _, ok := env["name"]; !ok {
		if machineID := system.MachineID(root); machineID != "" {
			env["name"] = machineID
		} else if hostname, err := system.Hostname(); err == nil {
			env["name"] = hostname
		} else {
			return errors.New("Unable to determine default etcd name")
		}
	}

	file := system.File{
		Path: path.Join(root, "run", "systemd", "system", "etcd.service.d", "20-cloudinit.conf"),
		RawFilePermissions: "0644",
		Content: env.String(),
	}

	return system.WriteFile(&file)
}
