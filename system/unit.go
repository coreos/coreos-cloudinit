/*
   Copyright 2014 CoreOS, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

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
