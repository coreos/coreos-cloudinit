package system

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestWriteFileUnencodedContent(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	fn := "foo"
	fullPath := path.Join(dir, fn)

	wf := File{
		Path:               fn,
		Content:            "bar",
		RawFilePermissions: "0644",
	}

	path, err := WriteFile(&wf, dir)
	if err != nil {
		t.Fatalf("Processing of WriteFile failed: %v", err)
	} else if path != fullPath {
		t.Fatalf("WriteFile returned bad path: want %s, got %s", fullPath, path)
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
	defer os.RemoveAll(dir)

	wf := File{
		Path:               path.Join(dir, "tmp", "foo"),
		Content:            "bar",
		RawFilePermissions: "pants",
	}

	if _, err := WriteFile(&wf, dir); err == nil {
		t.Fatalf("Expected error to be raised when writing file with invalid permission")
	}
}

func TestWriteFilePermissions(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	fn := "foo"
	fullPath := path.Join(dir, fn)

	wf := File{
		Path:               fn,
		RawFilePermissions: "0755",
	}

	path, err := WriteFile(&wf, dir)
	if err != nil {
		t.Fatalf("Processing of WriteFile failed: %v", err)
	} else if path != fullPath {
		t.Fatalf("WriteFile returned bad path: want %s, got %s", fullPath, path)
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
	defer os.RemoveAll(dir)

	wf := File{
		Path:     path.Join(dir, "tmp", "foo"),
		Content:  "",
		Encoding: "base64",
	}

	if _, err := WriteFile(&wf, dir); err == nil {
		t.Fatalf("Expected error to be raised when writing file with encoding")
	}
}
