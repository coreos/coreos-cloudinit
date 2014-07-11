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

	sd := &systemd{dir}

	dst := u.Destination(dir)
	expectDst := path.Join(dir, "run", "systemd", "network", "50-eth0.network")
	if dst != expectDst {
		t.Fatalf("unit.Destination returned %s, expected %s", dst, expectDst)
	}

	if err := sd.PlaceUnit(&u, dst); err != nil {
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

func TestUnitDestination(t *testing.T) {
	dir := "/some/dir"
	name := "foobar.service"

	u := Unit{
		Name:   name,
		DropIn: false,
	}

	dst := u.Destination(dir)
	expectDst := path.Join(dir, "etc", "systemd", "system", "foobar.service")
	if dst != expectDst {
		t.Errorf("unit.Destination returned %s, expected %s", dst, expectDst)
	}

	u.DropIn = true

	dst = u.Destination(dir)
	expectDst = path.Join(dir, "etc", "systemd", "system", "foobar.service.d", cloudConfigDropIn)
	if dst != expectDst {
		t.Errorf("unit.Destination returned %s, expected %s", dst, expectDst)
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

	sd := &systemd{dir}

	dst := u.Destination(dir)
	expectDst := path.Join(dir, "etc", "systemd", "system", "media-state.mount")
	if dst != expectDst {
		t.Fatalf("unit.Destination returned %s, expected %s", dst, expectDst)
	}

	if err := sd.PlaceUnit(&u, dst); err != nil {
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

	sd := &systemd{dir}

	// Ensure mask works with units that do not currently exist
	uf := &Unit{Name: "foo.service"}
	if err := sd.MaskUnit(uf); err != nil {
		t.Fatalf("Unable to mask new unit: %v", err)
	}
	fooPath := path.Join(dir, "etc", "systemd", "system", "foo.service")
	fooTgt, err := os.Readlink(fooPath)
	if err != nil {
		t.Fatalf("Unable to read link", err)
	}
	if fooTgt != "/dev/null" {
		t.Fatalf("unit not masked, got unit target", fooTgt)
	}

	// Ensure mask works with unit files that already exist
	ub := &Unit{Name: "bar.service"}
	barPath := path.Join(dir, "etc", "systemd", "system", "bar.service")
	if _, err := os.Create(barPath); err != nil {
		t.Fatalf("Error creating new unit file: %v", err)
	}
	if err := sd.MaskUnit(ub); err != nil {
		t.Fatalf("Unable to mask existing unit: %v", err)
	}
	barTgt, err := os.Readlink(barPath)
	if err != nil {
		t.Fatalf("Unable to read link", err)
	}
	if barTgt != "/dev/null" {
		t.Fatalf("unit not masked, got unit target", barTgt)
	}
}

func TestUnmaskUnit(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	sd := &systemd{dir}

	nilUnit := &Unit{Name: "null.service"}
	if err := sd.UnmaskUnit(nilUnit); err != nil {
		t.Errorf("unexpected error from unmasking nonexistent unit: %v", err)
	}

	uf := &Unit{Name: "foo.service", Content: "[Service]\nExecStart=/bin/true"}
	dst := uf.Destination(dir)
	if err := os.MkdirAll(path.Dir(dst), os.FileMode(0755)); err != nil {
		t.Fatalf("Unable to create unit directory: %v", err)
	}
	if _, err := os.Create(dst); err != nil {
		t.Fatalf("Unable to write unit file: %v", err)
	}

	if err := ioutil.WriteFile(dst, []byte(uf.Content), 700); err != nil {
		t.Fatalf("Unable to write unit file: %v", err)
	}
	if err := sd.UnmaskUnit(uf); err != nil {
		t.Errorf("unmask of non-empty unit returned unexpected error: %v", err)
	}
	got, _ := ioutil.ReadFile(dst)
	if string(got) != uf.Content {
		t.Errorf("unmask of non-empty unit mutated unit contents unexpectedly")
	}

	ub := &Unit{Name: "bar.service"}
	dst = ub.Destination(dir)
	if err := os.Symlink("/dev/null", dst); err != nil {
		t.Fatalf("Unable to create masked unit: %v", err)
	}
	if err := sd.UnmaskUnit(ub); err != nil {
		t.Errorf("unmask of unit returned unexpected error: %v", err)
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Errorf("expected %s to not exist after unmask, but got err: %s", err)
	}
}

func TestNullOrEmpty(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	non := path.Join(dir, "does_not_exist")
	ne, err := nullOrEmpty(non)
	if !os.IsNotExist(err) {
		t.Errorf("nullOrEmpty on nonexistent file returned bad error: %v", err)
	}
	if ne {
		t.Errorf("nullOrEmpty returned true unxpectedly")
	}

	regEmpty := path.Join(dir, "regular_empty_file")
	_, err = os.Create(regEmpty)
	if err != nil {
		t.Fatalf("Unable to create tempfile: %v", err)
	}
	gotNe, gotErr := nullOrEmpty(regEmpty)
	if !gotNe || gotErr != nil {
		t.Errorf("nullOrEmpty of regular empty file returned %t, %v - want true, nil", gotNe, gotErr)
	}

	reg := path.Join(dir, "regular_file")
	if err := ioutil.WriteFile(reg, []byte("asdf"), 700); err != nil {
		t.Fatalf("Unable to create tempfile: %v", err)
	}
	gotNe, gotErr = nullOrEmpty(reg)
	if gotNe || gotErr != nil {
		t.Errorf("nullOrEmpty of regular file returned %t, %v - want false, nil", gotNe, gotErr)
	}

	null := path.Join(dir, "null")
	if err := os.Symlink(os.DevNull, null); err != nil {
		t.Fatalf("Unable to create /dev/null link: %s", err)
	}
	gotNe, gotErr = nullOrEmpty(null)
	if !gotNe || gotErr != nil {
		t.Errorf("nullOrEmpty of null symlink returned %t, %v - want true, nil", gotNe, gotErr)
	}

}
