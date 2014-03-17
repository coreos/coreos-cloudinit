package cloudinit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type EtcdEnvironment map[string]string

func (ec EtcdEnvironment) String() (out string) {
	public := os.Getenv("COREOS_PUBLIC_IPV4")
	private := os.Getenv("COREOS_PRIVATE_IPV4")

	out += "[Service]\n"

	for key, val := range ec {
		key = strings.ToUpper(key)
		key = strings.Replace(key, "-", "_", -1)

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
func WriteEtcdEnvironment(root string, env EtcdEnvironment) error {
	cfgDir := path.Join(root, "etc", "systemd", "system", "etcd.service.d")
	cfgFile := path.Join(cfgDir, "20-cloudinit.conf")

	if _, err := os.Stat(cfgDir); err != nil {
		if err := os.MkdirAll(cfgDir, os.FileMode(0755)); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(cfgFile, []byte(env.String()), os.FileMode(0644))
}
