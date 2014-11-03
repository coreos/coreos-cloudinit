package initialize

import (
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/coreos/coreos-cloudinit/system"
)

const (
	base = `SERVER=https://example.com
GROUP=thegroupc`
	configured = base + `
REBOOT_STRATEGY=awesome
`
	expected = base + `
REBOOT_STRATEGY=etcd-lock
`
)

func setupFixtures(dir string) {
	os.MkdirAll(path.Join(dir, "usr", "share", "coreos"), 0755)
	os.MkdirAll(path.Join(dir, "run", "systemd", "system"), 0755)

	ioutil.WriteFile(path.Join(dir, "usr", "share", "coreos", "update.conf"), []byte(base), 0644)
}

func TestEmptyUpdateConfig(t *testing.T) {
	uc := &UpdateConfig{}
	f, err := uc.File("")
	if err != nil {
		t.Error("unexpected error getting file from empty UpdateConfig")
	}
	if f != nil {
		t.Errorf("getting file from empty UpdateConfig should have returned nil, got %v", f)
	}
	uu, err := uc.Units("")
	if err != nil {
		t.Error("unexpected error getting unit from empty UpdateConfig")
	}
	if len(uu) != 0 {
		t.Errorf("getting unit from empty UpdateConfig should have returned zero units, got %d", len(uu))
	}
}

func TestInvalidUpdateOptions(t *testing.T) {
	uon := &updateOption{
		key:    "numbers",
		prefix: "numero_",
		valid:  []string{"one", "two"},
	}
	uoa := &updateOption{
		key:    "any_will_do",
		prefix: "any_",
	}

	if !uon.isValid("one") {
		t.Error("update option did not accept valid option \"one\"")
	}
	if uon.isValid("three") {
		t.Error("update option accepted invalid option \"three\"")
	}
	for _, s := range []string{"one", "asdf", "foobarbaz"} {
		if !uoa.isValid(s) {
			t.Errorf("update option with no \"valid\" field did not accept %q", s)
		}
	}

	uc := &UpdateConfig{"reboot-strategy": "wizzlewazzle"}
	f, err := uc.File("")
	if err == nil {
		t.Errorf("File did not give an error on invalid UpdateOption")
	}
	if f != nil {
		t.Errorf("File did not return a nil file on invalid UpdateOption")
	}
}

func TestServerGroupOptions(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)
	setupFixtures(dir)
	u := &UpdateConfig{"group": "master", "server": "http://foo.com"}

	want := `
GROUP=master
SERVER=http://foo.com`

	f, err := u.File(dir)
	if err != nil {
		t.Errorf("unexpected error getting file from UpdateConfig: %v", err)
	} else if f == nil {
		t.Error("unexpectedly got empty file from UpdateConfig")
	} else {
		out := strings.Split(f.Content, "\n")
		sort.Strings(out)
		got := strings.Join(out, "\n")
		if got != want {
			t.Errorf("File has incorrect contents, got %v, want %v", got, want)
		}
	}

	uu, err := u.Units(dir)
	if err != nil {
		t.Errorf("unexpected error getting units from UpdateConfig: %v", err)
	} else if len(uu) != 1 {
		t.Errorf("unexpected number of files returned from UpdateConfig: want 1, got %d", len(uu))
	} else {
		unit := uu[0]
		if unit.Name != "update-engine.service" {
			t.Errorf("bad name for generated unit: want update-engine.service, got %s", unit.Name)
		}
		if unit.Command != "restart" {
			t.Errorf("bad command for generated unit: want restart, got %s", unit.Command)
		}
	}
}

func TestRebootStrategies(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)
	setupFixtures(dir)
	strategies := []struct {
		name     string
		line     string
		uMask    bool
		uCommand string
	}{
		{"best-effort", "REBOOT_STRATEGY=best-effort", false, "restart"},
		{"etcd-lock", "REBOOT_STRATEGY=etcd-lock", false, "restart"},
		{"reboot", "REBOOT_STRATEGY=reboot", false, "restart"},
		{"off", "REBOOT_STRATEGY=off", true, "stop"},
	}
	for _, s := range strategies {
		uc := &UpdateConfig{"reboot-strategy": s.name}
		f, err := uc.File(dir)
		if err != nil {
			t.Errorf("update failed to generate file for reboot-strategy=%v: %v", s.name, err)
		} else if f == nil {
			t.Errorf("generated empty file for reboot-strategy=%v", s.name)
		} else {
			seen := false
			for _, line := range strings.Split(f.Content, "\n") {
				if line == s.line {
					seen = true
					break
				}
			}
			if !seen {
				t.Errorf("couldn't find expected line %v for reboot-strategy=%v", s.line, s.name)
			}
		}
		uu, err := uc.Units(dir)
		if err != nil {
			t.Errorf("failed to generate unit for reboot-strategy=%v!", s.name)
		} else if len(uu) != 1 {
			t.Errorf("unexpected number of units for reboot-strategy=%v: %d", s.name, len(uu))
		} else {
			u := uu[0]
			if u.Name != locksmithUnit {
				t.Errorf("unit generated for reboot strategy=%v had bad name: %v", s.name, u.Name)
			}
			if u.Mask != s.uMask {
				t.Errorf("unit generated for reboot strategy=%v had bad mask: %t", s.name, u.Mask)
			}
			if u.Command != s.uCommand {
				t.Errorf("unit generated for reboot strategy=%v had bad command: %v", s.name, u.Command)
			}
		}
	}

}

func TestUpdateConfWrittenToDisk(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)
	setupFixtures(dir)

	for i := 0; i < 2; i++ {
		if i == 1 {
			err = ioutil.WriteFile(path.Join(dir, "etc", "coreos", "update.conf"), []byte(configured), 0644)
			if err != nil {
				t.Fatal(err)
			}
		}
		uc := &UpdateConfig{"reboot-strategy": "etcd-lock"}

		f, err := uc.File(dir)
		if err != nil {
			t.Fatalf("Processing UpdateConfig failed: %v", err)
		} else if f == nil {
			t.Fatal("Unexpectedly got nil updateconfig file")
		}

		if _, err := system.WriteFile(f, dir); err != nil {
			t.Fatalf("Error writing update config: %v", err)
		}

		fullPath := path.Join(dir, "etc", "coreos", "update.conf")

		fi, err := os.Stat(fullPath)
		if err != nil {
			t.Fatalf("Unable to stat file: %v", err)
		}

		if fi.Mode() != os.FileMode(0644) {
			t.Errorf("File has incorrect mode: %v", fi.Mode())
		}

		contents, err := ioutil.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("Unable to read expected file: %v", err)
		}

		if string(contents) != expected {
			t.Fatalf("File has incorrect contents, got %v, wanted %v", string(contents), expected)
		}
	}
}
