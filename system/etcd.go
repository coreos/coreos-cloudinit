package system

import (
	"github.com/coreos/coreos-cloudinit/config"
)

// Etcd is a top-level structure which embeds its underlying configuration,
// config.Etcd, and provides the system-specific Unit().
type Etcd struct {
	config.Etcd
}

// Units creates a Unit file drop-in for etcd, using any configured options.
func (ee Etcd) Units() []Unit {
	content := dropinContents(ee.Etcd)
	if content == "" {
		return nil
	}
	return []Unit{{config.Unit{
		Name:    "etcd.service",
		Runtime: true,
		DropIn:  true,
		Content: content,
	}}}
}
