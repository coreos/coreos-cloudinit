package system

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
)

type File struct {
	Encoding           string
	Content            string
	Owner              string
	Path               string
	RawFilePermissions string `yaml:"permissions"`
}

func (f *File) Permissions() (os.FileMode, error) {
	if f.RawFilePermissions == "" {
		return os.FileMode(0644), nil
	}

	// Parse string representation of file mode as octal
	perm, err := strconv.ParseInt(f.RawFilePermissions, 8, 32)
	if err != nil {
		return 0, errors.New("Unable to parse file permissions as octal integer")
	}
	return os.FileMode(perm), nil
}

func WriteFile(f *File, root string) (string, error) {
	if f.Encoding != "" {
		return "", fmt.Errorf("Unable to write file with encoding %s", f.Encoding)
	}

	fullpath := path.Join(root, f.Path)
	dir := path.Dir(fullpath)

	if err := EnsureDirectoryExists(dir); err != nil {
		return "", err
	}

	perm, err := f.Permissions()
	if err != nil {
		return "", err
	}

	var tmp *os.File
	// Create a temporary file in the same directory to ensure it's on the same filesystem
	if tmp, err = ioutil.TempFile(dir, "cloudinit-temp"); err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(tmp.Name(), []byte(f.Content), perm); err != nil {
		return "", err
	}

	if err := tmp.Close(); err != nil {
		return "", err
	}

	// Ensure the permissions are as requested (since WriteFile can be affected by sticky bit)
	if err := os.Chmod(tmp.Name(), perm); err != nil {
		return "", err
	}

	if f.Owner != "" {
		// We shell out since we don't have a way to look up unix groups natively
		cmd := exec.Command("chown", f.Owner, tmp.Name())
		if err := cmd.Run(); err != nil {
			return "", err
		}
	}

	if err := os.Rename(tmp.Name(), fullpath); err != nil {
		return "", err
	}

	return fullpath, nil
}

func EnsureDirectoryExists(dir string) error {
	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", dir)
		}
	} else {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
