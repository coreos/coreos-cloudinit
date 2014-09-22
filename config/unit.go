package config

import (
	"path/filepath"
	"strings"
)

type Unit struct {
	Name    string `yaml:"name"`
	Mask    bool   `yaml:"mask"`
	Enable  bool   `yaml:"enable"`
	Runtime bool   `yaml:"runtime"`
	Content string `yaml:"content"`
	Command string `yaml:"command"`

	// For drop-in units, a cloudinit.conf is generated.
	// This is currently unbound in YAML (and hence unsettable in cloud-config files)
	// until the correct behaviour for multiple drop-in units is determined.
	DropIn bool `yaml:"-"`
}

func (u *Unit) Type() string {
	ext := filepath.Ext(u.Name)
	return strings.TrimLeft(ext, ".")
}

func (u *Unit) Group() string {
	switch u.Type() {
	case "network", "netdev", "link":
		return "network"
	default:
		return "system"
	}
}
