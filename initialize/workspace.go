package initialize

import (
	"io/ioutil"
	"path"

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

	file := system.File{
		Path: tmp.Name(),
		RawFilePermissions: "0744",
		Content: string(script),
	}

	err = system.WriteFile(&file)
	return file.Path, err
}

func PersistUnitNameInWorkspace(name string, workspace string) error {
	file := system.File{
		Path: path.Join(workspace, "scripts", "unit-name"),
		RawFilePermissions: "0644",
		Content: name,
	}
	return system.WriteFile(&file)
}
