package system

import (
	"io/ioutil"
	"os"
	"path"
	"syscall"
	"testing"
)

func TestPlaceNetworkUnit(t *testing.T) {
	u := Unit{
		Name: "50-eth0.network",
      Runtime: true,
      Content: `[Match]
Name=eth47

[Network]
Address=10.209.171.177/19
`,
	}

	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer syscall.Rmdir(dir)

	if _, err := PlaceUnit(&u, dir); err != nil {
		t.Fatalf("PlaceUnit failed: %v", err)
	}

	fullPath := path.Join(dir, "run", "systemd", "network", "50-eth0.network")
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

	expect := `[Match]
Name=eth47

[Network]
Address=10.209.171.177/19
`
	if string(contents) != expect {
		t.Fatalf("File has incorrect contents '%s'.\nExpected '%s'", string(contents), expect)
	}
}

func TestPlaceMountUnit(t *testing.T) {
	u := Unit{
		Name: "media-state.mount",
      Runtime: false,
      Content: `[Mount]
What=/dev/sdb1
Where=/media/state
`,
	}

	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer syscall.Rmdir(dir)

	if _, err := PlaceUnit(&u, dir); err != nil {
		t.Fatalf("PlaceUnit failed: %v", err)
	}

	fullPath := path.Join(dir, "etc", "systemd", "system", "media-state.mount")
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

	expect := `[Mount]
What=/dev/sdb1
Where=/media/state
`
	if string(contents) != expect {
		t.Fatalf("File has incorrect contents '%s'.\nExpected '%s'", string(contents), expect)
	}
}

func TestMachineID(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer syscall.Rmdir(dir)

	os.Mkdir(path.Join(dir, "etc"), os.FileMode(0755))
	ioutil.WriteFile(path.Join(dir, "etc", "machine-id"), []byte("node007\n"), os.FileMode(0444))

	if MachineID(dir) != "node007" {
		t.Fatalf("File has incorrect contents")
	}
}
