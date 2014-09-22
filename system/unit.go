package system

import (
	"fmt"
	"path"

	"github.com/coreos/coreos-cloudinit/config"
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

// Unit is a top-level structure which embeds its underlying configuration,
// config.Unit, and provides the system-specific Destination().
type Unit struct {
	config.Unit
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
