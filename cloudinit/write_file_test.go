package cloudinit

import (
	"io/ioutil"
	"os"
	"path"
	"syscall"
	"testing"
)

func TestWriteFileUnencodedContent(t *testing.T) {
	wf := WriteFile{
		Path:        "/tmp/foo",
		Content:     "bar",
		Permissions: "0644",
	}
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer syscall.Rmdir(dir)

	if err := ProcessWriteFile(dir, &wf); err != nil {
		t.Fatalf("Processing of WriteFile failed: %v", err)
	}

	fullPath := path.Join(dir, "tmp", "foo")

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

	if string(contents) != "bar" {
		t.Fatalf("File has incorrect contents")
	}
}

func TestWriteFileInvalidPermission(t *testing.T) {
	wf := WriteFile{
		Path:        "/tmp/foo",
		Content:     "bar",
		Permissions: "pants",
	}
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer syscall.Rmdir(dir)

	if err := ProcessWriteFile(dir, &wf); err == nil {
		t.Fatalf("Expected error to be raised when writing file with invalid permission")
	}
}

func TestWriteFileEncodedContent(t *testing.T) {
	wf := WriteFile{
		Path: "/tmp/foo",
		Content: "",
		Encoding: "base64",
	}

	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer syscall.Rmdir(dir)

	if err := ProcessWriteFile(dir, &wf); err == nil {
		t.Fatalf("Expected error to be raised when writing file with encoding")
	}
}
