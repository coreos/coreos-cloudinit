package initialize

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/coreos/coreos-cloudinit/system"
)

func PrepWorkspace(workspace string) error {
	if err := system.EnsureDirectoryExists(workspace); err != nil {
		return err
	}

	scripts := path.Join(workspace, "scripts")
	if err := system.EnsureDirectoryExists(scripts); err != nil {
		return err
	}

	return nil
}

func PersistScriptInWorkspace(script system.Script, workspace string) (string, error) {
	scriptsPath := path.Join(workspace, "scripts")
	tmp, err := ioutil.TempFile(scriptsPath, "")
	if err != nil {
		return "", err
	}
	tmp.Close()

	relpath := strings.TrimPrefix(tmp.Name(), workspace)

	file := system.File{
		Path:               relpath,
		RawFilePermissions: "0744",
		Content:            string(script),
	}

	return system.WriteFile(&file, workspace)
}

func PersistUnitNameInWorkspace(name string, workspace string) error {
	file := system.File{
		Path:               path.Join("scripts", "unit-name"),
		RawFilePermissions: "0644",
		Content:            name,
	}
	_, err := system.WriteFile(&file, workspace)
	return err
}
