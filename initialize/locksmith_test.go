package initialize

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
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
	os.MkdirAll(path.Join(dir, "etc", "coreos"), 0755)
	os.MkdirAll(path.Join(dir, "run", "systemd", "system"), 0755)

	ioutil.WriteFile(path.Join(dir, "usr", "share", "coreos", "update.conf"), []byte(base), 0644)
}

func TestLocksmithEnvironmentWrittenToDisk(t *testing.T) {
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

		if err := WriteLocksmithConfig("etcd-lock", dir); err != nil {
			t.Fatalf("Processing of LocksmithEnvironment failed: %v", err)
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
func TestLocksmithEnvironmentMasked(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)
	setupFixtures(dir)

	if err := WriteLocksmithConfig("off", dir); err != nil {
		t.Fatalf("Processing of LocksmithEnvironment failed: %v", err)
	}

	fullPath := path.Join(dir, "etc", "systemd", "system", "locksmithd.service")
	target, err := os.Readlink(fullPath)
	if err != nil {
		t.Fatalf("Unable to read link %v", err)
	}
	if target != "/dev/null" {
		t.Fatalf("Locksmith not masked, unit target %v", target)
	}
}
