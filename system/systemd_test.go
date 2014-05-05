package system

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestPlaceNetworkUnit(t *testing.T) {
	u := Unit{
		Name:    "50-eth0.network",
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
	defer os.RemoveAll(dir)

	dst := UnitDestination(&u, dir)
	expectDst := path.Join(dir, "run", "systemd", "network", "50-eth0.network")
	if dst != expectDst {
		t.Fatalf("UnitDestination returned %s, expected %s", dst, expectDst)
	}

	if err := PlaceUnit(&u, dst); err != nil {
		t.Fatalf("PlaceUnit failed: %v", err)
	}

	fi, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if fi.Mode() != os.FileMode(0644) {
		t.Errorf("File has incorrect mode: %v", fi.Mode())
	}

	contents, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	expectContents := `[Match]
Name=eth47

[Network]
Address=10.209.171.177/19
`
	if string(contents) != expectContents {
		t.Fatalf("File has incorrect contents '%s'.\nExpected '%s'", string(contents), expectContents)
	}
}

func TestPlaceMountUnit(t *testing.T) {
	u := Unit{
		Name:    "media-state.mount",
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
	defer os.RemoveAll(dir)

	dst := UnitDestination(&u, dir)
	expectDst := path.Join(dir, "etc", "systemd", "system", "media-state.mount")
	if dst != expectDst {
		t.Fatalf("UnitDestination returned %s, expected %s", dst, expectDst)
	}

	if err := PlaceUnit(&u, dst); err != nil {
		t.Fatalf("PlaceUnit failed: %v", err)
	}

	fi, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if fi.Mode() != os.FileMode(0644) {
		t.Errorf("File has incorrect mode: %v", fi.Mode())
	}

	contents, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	expectContents := `[Mount]
What=/dev/sdb1
Where=/media/state
`
	if string(contents) != expectContents {
		t.Fatalf("File has incorrect contents '%s'.\nExpected '%s'", string(contents), expectContents)
	}
}

func TestMachineID(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	os.Mkdir(path.Join(dir, "etc"), os.FileMode(0755))
	ioutil.WriteFile(path.Join(dir, "etc", "machine-id"), []byte("node007\n"), os.FileMode(0444))

	if MachineID(dir) != "node007" {
		t.Fatalf("File has incorrect contents")
	}
}
func TestMaskUnit(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)
	if err := MaskUnit("foo.service", dir); err != nil {
		t.Fatalf("Unable to mask unit: %v", err)
	}

	fullPath := path.Join(dir, "etc", "systemd", "system", "foo.service")
	target, err := os.Readlink(fullPath)
	if err != nil {
		t.Fatalf("Unable to read link", err)
	}
	if target != "/dev/null" {
		t.Fatalf("unit not masked, got unit target", target)
	}
}
