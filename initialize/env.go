package initialize

import (
	"os"
	"path"
	"strings"
)

const DefaultSSHKeyName = "coreos-cloudinit"

type Environment struct {
	root          string
	configRoot    string
	workspace     string
	netconfType   string
	sshKeyName    string
	substitutions map[string]string
}

func NewEnvironment(root, configRoot, workspace, netconfType, sshKeyName string) *Environment {
	substitutions := map[string]string{
		"$public_ipv4":  os.Getenv("COREOS_PUBLIC_IPV4"),
		"$private_ipv4": os.Getenv("COREOS_PRIVATE_IPV4"),
	}
	return &Environment{root, configRoot, workspace, netconfType, sshKeyName, substitutions}
}

func (self *Environment) Workspace() string {
	return path.Join(self.root, self.workspace)
}

func (self *Environment) Root() string {
	return self.root
}

func (self *Environment) ConfigRoot() string {
	return self.configRoot
}

func (self *Environment) NetconfType() string {
	return self.netconfType
}

func (self *Environment) SSHKeyName() string {
	return self.sshKeyName
}

func (self *Environment) SetSSHKeyName(name string) {
	self.sshKeyName = name
}

func (self *Environment) Apply(data string) string {
	for key, val := range self.substitutions {
		data = strings.Replace(data, key, val, -1)
	}
	return data
}

// normalizeSvcEnv standardizes the keys of the map (environment variables for a service)
// by replacing any dashes with underscores and ensuring they are entirely upper case.
// For example, "some-env" --> "SOME_ENV"
func normalizeSvcEnv(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for key, val := range m {
		key = strings.ToUpper(key)
		key = strings.Replace(key, "-", "_", -1)
		out[key] = val
	}
	return out
}
