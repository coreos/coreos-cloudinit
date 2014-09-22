package system

import (
	"github.com/coreos/coreos-cloudinit/config"
)

// Fleet is a top-level structure which embeds its underlying configuration,
// config.Fleet, and provides the system-specific Unit().
type Fleet struct {
	config.Fleet
}

// Units generates a Unit file drop-in for fleet, if any fleet options were
// configured in cloud-config
func (fe Fleet) Units() ([]Unit, error) {
	content := dropinContents(fe.Fleet)
	if content == "" {
		return nil, nil
	}
	return []Unit{{
		Name:    "fleet.service",
		Runtime: true,
		DropIn:  true,
		Content: content,
	}}, nil
}
