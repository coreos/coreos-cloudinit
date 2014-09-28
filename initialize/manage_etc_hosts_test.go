package initialize

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/coreos/coreos-cloudinit/system"
)

func TestCloudConfigManageEtcHosts(t *testing.T) {
	contents := `
manage_etc_hosts: localhost
`
	cfg, err := NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error: %v", err)
	}

	manageEtcHosts := cfg.ManageEtcHosts

	if manageEtcHosts != "localhost" {
		t.Errorf("ManageEtcHosts value is %q, expected 'localhost'", manageEtcHosts)
	}
}

func TestManageEtcHostsInvalidValue(t *testing.T) {
	eh := EtcHosts("invalid")
	if f, err := eh.File(""); err == nil || f != nil {
		t.Fatalf("EtcHosts File succeeded with invalid value!")
	}
}

func TestEtcHostsWrittenToDisk(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	eh := EtcHosts("localhost")

	f, err := eh.File(dir)
	if err != nil {
		t.Fatalf("Error calling File on EtcHosts: %v", err)
	}
	if f == nil {
		t.Fatalf("manageEtcHosts returned nil file unexpectedly")
	}

	if _, err := system.WriteFile(f, dir); err != nil {
		t.Fatalf("Error writing EtcHosts: %v", err)
	}

	fullPath := path.Join(dir, "etc", "hosts")

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

	hostname, err := os.Hostname()
	if err != nil {
		t.Fatalf("Unable to read OS hostname: %v", err)
	}

	expect := fmt.Sprintf("%s %s\n", DefaultIpv4Address, hostname)

	if string(contents) != expect {
		t.Fatalf("File has incorrect contents")
	}
}
