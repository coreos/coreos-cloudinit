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

// TODO(jonboulle): this is getting unwieldy, should be able to simplify the interface somehow
func NewEnvironment(root, configRoot, workspace, netconfType, sshKeyName string, substitutions map[string]string) *Environment {
	// If certain values are not in the supplied substitution, fall back to retrieving them from the environment
	for k, v := range map[string]string{
		"$public_ipv4":  os.Getenv("COREOS_PUBLIC_IPV4"),
		"$private_ipv4": os.Getenv("COREOS_PRIVATE_IPV4"),
	} {
		if _, ok := substitutions[k]; !ok {
			substitutions[k] = v
		}
	}
	return &Environment{root, configRoot, workspace, netconfType, sshKeyName, substitutions}
}

func (e *Environment) Workspace() string {
	return path.Join(e.root, e.workspace)
}

func (e *Environment) Root() string {
	return e.root
}

func (e *Environment) ConfigRoot() string {
	return e.configRoot
}

func (e *Environment) NetconfType() string {
	return e.netconfType
}

func (e *Environment) SSHKeyName() string {
	return e.sshKeyName
}

func (e *Environment) SetSSHKeyName(name string) {
	e.sshKeyName = name
}

func (e *Environment) Apply(data string) string {
	for key, val := range e.substitutions {
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
