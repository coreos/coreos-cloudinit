package system

import (
	"io/ioutil"
	"os"
	"path"
	"syscall"
	"testing"
)

func TestWriteFileUnencodedContent(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer syscall.Rmdir(dir)

	fullPath := path.Join(dir, "tmp", "foo")

	wf := File{
		Path:        fullPath,
		Content:     "bar",
		RawFilePermissions: "0644",
	}

	if err := WriteFile(&wf); err != nil {
		t.Fatalf("Processing of WriteFile failed: %v", err)
	}

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
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer syscall.Rmdir(dir)

	wf := File{
		Path:        path.Join(dir, "tmp", "foo"),
		Content:     "bar",
		RawFilePermissions: "pants",
	}

	if err := WriteFile(&wf); err == nil {
		t.Fatalf("Expected error to be raised when writing file with invalid permission")
	}
}

func TestWriteFilePermissions(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer syscall.Rmdir(dir)

	fullPath := path.Join(dir, "tmp", "foo")

	wf := File{
		Path:               fullPath,
		RawFilePermissions: "0755",
	}

	if err := WriteFile(&wf); err != nil {
		t.Fatalf("Processing of WriteFile failed: %v", err)
	}

	fi, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if fi.Mode() != os.FileMode(0755) {
		t.Errorf("File has incorrect mode: %v", fi.Mode())
	}
}

func TestWriteFileEncodedContent(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer syscall.Rmdir(dir)

	wf := File{
		Path: path.Join(dir, "tmp", "foo"),
		Content: "",
		Encoding: "base64",
	}

	if err := WriteFile(&wf); err == nil {
		t.Fatalf("Expected error to be raised when writing file with encoding")
	}
}
