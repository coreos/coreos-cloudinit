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

package initialize

import (
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/system"
)

type TestUnitManager struct {
	placed   []string
	enabled  []string
	masked   []string
	unmasked []string
	commands map[string]string
	reload   bool
}

func (tum *TestUnitManager) PlaceUnit(unit *system.Unit, dst string) error {
	tum.placed = append(tum.placed, unit.Name)
	return nil
}
func (tum *TestUnitManager) EnableUnitFile(unit string, runtime bool) error {
	tum.enabled = append(tum.enabled, unit)
	return nil
}
func (tum *TestUnitManager) RunUnitCommand(command, unit string) (string, error) {
	tum.commands = make(map[string]string)
	tum.commands[unit] = command
	return "", nil
}
func (tum *TestUnitManager) DaemonReload() error {
	tum.reload = true
	return nil
}
func (tum *TestUnitManager) MaskUnit(unit *system.Unit) error {
	tum.masked = append(tum.masked, unit.Name)
	return nil
}
func (tum *TestUnitManager) UnmaskUnit(unit *system.Unit) error {
	tum.unmasked = append(tum.unmasked, unit.Name)
	return nil
}

func TestProcessUnits(t *testing.T) {
	tum := &TestUnitManager{}
	units := []system.Unit{
		system.Unit{config.Unit{
			Name: "foo",
			Mask: true,
		}},
	}
	if err := processUnits(units, "", tum); err != nil {
		t.Fatalf("unexpected error calling processUnits: %v", err)
	}
	if len(tum.masked) != 1 || tum.masked[0] != "foo" {
		t.Errorf("expected foo to be masked, but found %v", tum.masked)
	}

	tum = &TestUnitManager{}
	units = []system.Unit{
		system.Unit{config.Unit{
			Name: "bar.network",
		}},
	}
	if err := processUnits(units, "", tum); err != nil {
		t.Fatalf("unexpected error calling processUnits: %v", err)
	}
	if _, ok := tum.commands["systemd-networkd.service"]; !ok {
		t.Errorf("expected systemd-networkd.service to be reloaded!")
	}

	tum = &TestUnitManager{}
	units = []system.Unit{
		system.Unit{config.Unit{
			Name:    "baz.service",
			Content: "[Service]\nExecStart=/bin/true",
		}},
	}
	if err := processUnits(units, "", tum); err != nil {
		t.Fatalf("unexpected error calling processUnits: %v", err)
	}
	if len(tum.placed) != 1 || tum.placed[0] != "baz.service" {
		t.Fatalf("expected baz.service to be written, but got %v", tum.placed)
	}

	tum = &TestUnitManager{}
	units = []system.Unit{
		system.Unit{config.Unit{
			Name:    "locksmithd.service",
			Runtime: true,
		}},
	}
	if err := processUnits(units, "", tum); err != nil {
		t.Fatalf("unexpected error calling processUnits: %v", err)
	}
	if len(tum.unmasked) != 1 || tum.unmasked[0] != "locksmithd.service" {
		t.Fatalf("expected locksmithd.service to be unmasked, but got %v", tum.unmasked)
	}

	tum = &TestUnitManager{}
	units = []system.Unit{
		system.Unit{config.Unit{
			Name:   "woof",
			Enable: true,
		}},
	}
	if err := processUnits(units, "", tum); err != nil {
		t.Fatalf("unexpected error calling processUnits: %v", err)
	}
	if len(tum.enabled) != 1 || tum.enabled[0] != "woof" {
		t.Fatalf("expected woof to be enabled, but got %v", tum.enabled)
	}
}
