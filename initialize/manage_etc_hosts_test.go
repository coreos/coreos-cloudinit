package initialize

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
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
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer rmdir(dir)

	if err := WriteEtcHosts("invalid", dir); err == nil {
		t.Fatalf("WriteEtcHosts succeeded with invalid value: %v", err)
	}
}

func TestEtcHostsWrittenToDisk(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer rmdir(dir)

	if err := WriteEtcHosts("localhost", dir); err != nil {
		t.Fatalf("WriteEtcHosts failed: %v", err)
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
