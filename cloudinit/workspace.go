package cloudinit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

func PrepWorkspace(workspace string) error {
	// Ensure workspace exists and is a directory
	info, err := os.Stat(workspace)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", workspace)
		}
	} else {
		err = os.MkdirAll(workspace, 0755)
		if err != nil {
			return err
		}
	}

	// Ensure scripts dir in workspace exists and is a directory
	scripts := path.Join(workspace, "scripts")
	info, err = os.Stat(scripts)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", scripts)
		}
	} else {
		err = os.Mkdir(scripts, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func PersistScriptInWorkspace(script Script, workspace string) (string, error) {
	scriptsDir := path.Join(workspace, "scripts")
	f, err := ioutil.TempFile(scriptsDir, "")
	if err != nil {
		return "", err
	}
	defer f.Close()

	f.Chmod(0744)

	_, err = f.Write(script)
	if err != nil {
		return "", err
	}

	// Ensure script has been written to disk before returning, as the
	// next natural thing to do is execute it
	f.Sync()

	return f.Name(), nil
}

func PersistScriptUnitNameInWorkspace(name string, workspace string) error {
	unitPath := path.Join(workspace, "scripts", "unit-name")
	return ioutil.WriteFile(unitPath, []byte(name), 0644)
}
