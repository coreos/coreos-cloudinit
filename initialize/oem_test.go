package initialize

import (
	"io/ioutil"
	"os"
	"path"
	"syscall"
	"testing"
)

func TestOEMReleaseWrittenToDisk(t *testing.T) {
	oem := OEMRelease{
		ID:           "rackspace",
		Name:         "Rackspace Cloud Servers",
		VersionID:    "168.0.0",
		HomeURL:      "https://www.rackspace.com/cloud/servers/",
		BugReportURL: "https://github.com/coreos/coreos-overlay",
	}
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer syscall.Rmdir(dir)

	if err := WriteOEMRelease(&oem, dir); err != nil {
		t.Fatalf("Processing of EtcdEnvironment failed: %v", err)
	}

	fullPath := path.Join(dir, "etc", "oem-release")

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

	expect := `ID=rackspace
VERSION_ID=168.0.0
NAME="Rackspace Cloud Servers"
HOME_URL="https://www.rackspace.com/cloud/servers/"
BUG_REPORT_URL="https://github.com/coreos/coreos-overlay"
`
	if string(contents) != expect {
		t.Fatalf("File has incorrect contents")
	}
}
