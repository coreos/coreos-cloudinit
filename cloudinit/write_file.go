package cloudinit

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
)

type WriteFile struct {
	Encoding    string
	Content     string
	Owner       string
	Path        string
	Permissions string
}

func ProcessWriteFile(base string, wf *WriteFile) error {
	fullPath := path.Join(base, wf.Path)

	if err := os.MkdirAll(path.Dir(fullPath), os.FileMode(0744)); err != nil {
		return err
	}

	// Parse string representation of file mode as octal
	perm, err := strconv.ParseInt(wf.Permissions, 8, 32)
	if err != nil {
		return errors.New("Unable to parse file permissions as octal integer")
	}

	if err := ioutil.WriteFile(fullPath, []byte(wf.Content), os.FileMode(perm)); err != nil {
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
