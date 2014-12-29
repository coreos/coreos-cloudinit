package system

import (
	"path"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"
)

// flannel is a top-level structure which embeds its underlying configuration,
// config.Flannel, and provides the system-specific Unit().
type Flannel struct {
	config.Flannel
}

func (fl Flannel) envVars() string {
	return strings.Join(getEnvVars(fl.Flannel), "\n")
}

func (fl Flannel) File() (*File, error) {
	vars := fl.envVars()
	if vars == "" {
		return nil, nil
	}
	return &File{config.File{
		Path:               path.Join("run", "flannel", "options.env"),
		RawFilePermissions: "0644",
		Content:            vars,
	}}, nil
}
