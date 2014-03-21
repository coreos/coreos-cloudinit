package initialize

import (
	"os"
	"path"
	"strings"
)

const DefaultSSHKeyName = "coreos-cloudinit"

type Environment struct {
	root          string
	workspace     string
	sshKeyName    string
	substitutions map[string]string
}

func NewEnvironment(root, workspace string) *Environment {
	substitutions := map[string]string{
		"$public_ipv4":  os.Getenv("COREOS_PUBLIC_IPV4"),
		"$private_ipv4": os.Getenv("COREOS_PRIVATE_IPV4"),
	}
	return &Environment{root, workspace, DefaultSSHKeyName, substitutions}
}

func (self *Environment) Workspace() string {
	return path.Join(self.root, self.workspace)
}

func (self *Environment) Root() string {
	return self.root
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
