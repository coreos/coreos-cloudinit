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
	Encoding    string
	Content     string
	Owner       string
	Path        string
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


func WriteFile(f *File) error {
	if f.Encoding != "" {
		return fmt.Errorf("Unable to write file with encoding %s", f.Encoding)
	}

	if err := os.MkdirAll(path.Dir(f.Path), os.FileMode(0755)); err != nil {
		return err
	}

	perm, err := f.Permissions()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(f.Path, []byte(f.Content), perm); err != nil {
		return err
	}

	if f.Owner != "" {
		// We shell out since we don't have a way to look up unix groups natively
		cmd := exec.Command("chown", f.Owner, f.Path)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
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
