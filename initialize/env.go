package initialize

import (
	"path"
)

const DefaultSSHKeyName = "coreos-cloudinit"

type Environment struct {
	root      string
	workspace string
	sshKeyName string
}

func NewEnvironment(root, workspace string) *Environment {
	return &Environment{root, workspace, DefaultSSHKeyName}
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
