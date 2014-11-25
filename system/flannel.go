package system

import (
	"github.com/coreos/coreos-cloudinit/config"
)

// flannel is a top-level structure which embeds its underlying configuration,
// config.Flannel, and provides the system-specific Unit().
type Flannel struct {
	config.Flannel
}

// Units generates a Unit file drop-in for flannel, if any flannel options were
// configured in cloud-config
func (fl Flannel) Units() []Unit {
	return dropinFromConfig(fl.Flannel, "flanneld.service")
}
