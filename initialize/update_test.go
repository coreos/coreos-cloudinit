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
	u, err := uc.Unit("")
	if err != nil {
		t.Error("unexpected error getting unit from empty UpdateConfig")
	}
	if u != nil {
		t.Errorf("getting unit from empty UpdateConfig should have returned nil, got %v", u)
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
				t.Errorf("couldn't find expected line %v for reboot-strategy=%v", s.line)
			}
		}
		u, err := uc.Unit(dir)
		if err != nil {
			t.Errorf("failed to generate unit for reboot-strategy=%v!", s.name)
		} else if u == nil {
			t.Errorf("generated empty unit for reboot-strategy=%v", s.name)
		} else {
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

		f.Path = path.Join(dir, f.Path)
		if err := system.WriteFile(f); err != nil {
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
