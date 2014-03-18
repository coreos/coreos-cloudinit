package cloudinit

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
)

type WriteFile struct {
	Encoding           string
	Content            string
	Owner              string
	Path               string
	RawFilePermissions string `yaml:"permissions"`
}

func (wf *WriteFile) Permissions() (os.FileMode, error) {
	if wf.RawFilePermissions == "" {
		return os.FileMode(0644), nil
	}

	// Parse string representation of file mode as octal
	perm, err := strconv.ParseInt(wf.RawFilePermissions, 8, 32)
	if err != nil {
		return 0, errors.New("Unable to parse file permissions as octal integer")
	}
	return os.FileMode(perm), nil
}

func ProcessWriteFile(base string, wf *WriteFile) error {
	if wf.Encoding != "" {
		return fmt.Errorf("Unable to write file with encoding %s", wf.Encoding)
	}

	fullPath := path.Join(base, wf.Path)

	if err := os.MkdirAll(path.Dir(fullPath), os.FileMode(0744)); err != nil {
		return err
	}

	perm, err := wf.Permissions()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(fullPath, []byte(wf.Content), perm); err != nil {
		return err
	}

	if wf.Owner != "" {
		// We shell out since we don't have a way to look up unix groups natively
		cmd := exec.Command("chown", wf.Owner, fullPath)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
