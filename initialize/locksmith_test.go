package initialize

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestLocksmithEnvironmentWrittenToDisk(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := WriteLocksmithEnvironment("etcd-lock", dir); err != nil {
		t.Fatalf("Processing of LocksmithEnvironment failed: %v", err)
	}

	fullPath := path.Join(dir, "run", "systemd", "system", "locksmithd.service.d", "20-cloudinit.conf")

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

	expect := `[Service]
Environment=LOCKSMITH_STRATEGY=etcd-lock`
	if string(contents) != expect {
		t.Fatalf("File has incorrect contents, got %v, wanted %v", contents, expect)
	}
}
func TestLocksmithEnvironmentMasked(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := WriteLocksmithEnvironment("none", dir); err != nil {
		t.Fatalf("Processing of LocksmithEnvironment failed: %v", err)
	}

	fullPath := path.Join(dir, "run", "systemd", "system", "locksmithd.service")
	target, err := os.Readlink(fullPath)
	if err != nil {
		t.Fatalf("Unable to read link %v", err)
	}
	if target != "/dev/null" {
		t.Fatalf("Locksmith not masked, unit target %v", target)
	}
}
