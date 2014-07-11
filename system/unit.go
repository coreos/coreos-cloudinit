package system

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// Name for drop-in service configuration files created by cloudconfig
const cloudConfigDropIn = "20-cloudinit.conf"

type UnitManager interface {
	PlaceUnit(unit *Unit, dst string) error
	EnableUnitFile(unit string, runtime bool) error
	RunUnitCommand(command, unit string) (string, error)
	DaemonReload() error
	MaskUnit(unit *Unit) error
	UnmaskUnit(unit *Unit) error
}

type Unit struct {
	Name    string
	Mask    bool
	Enable  bool
	Runtime bool
	Content string
	Command string

	// For drop-in units, a cloudinit.conf is generated.
	// This is currently unbound in YAML (and hence unsettable in cloud-config files)
	// until the correct behaviour for multiple drop-in units is determined.
	DropIn bool `yaml:"-"`
}

func (u *Unit) Type() string {
	ext := filepath.Ext(u.Name)
	return strings.TrimLeft(ext, ".")
}

func (u *Unit) Group() (group string) {
	t := u.Type()
	if t == "network" || t == "netdev" || t == "link" {
		group = "network"
	} else {
		group = "system"
	}
	return
}

type Script []byte

// Destination builds the appropriate absolute file path for
// the Unit. The root argument indicates the effective base
// directory of the system (similar to a chroot).
func (u *Unit) Destination(root string) string {
	dir := "etc"
	if u.Runtime {
		dir = "run"
	}

	if u.DropIn {
		return path.Join(root, dir, "systemd", u.Group(), fmt.Sprintf("%s.d", u.Name), cloudConfigDropIn)
	} else {
		return path.Join(root, dir, "systemd", u.Group(), u.Name)
	}
}
