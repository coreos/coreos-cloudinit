package initialize

import (
	"fmt"
	"path"

	"github.com/coreos/coreos-cloudinit/system"
)

const locksmithUnit = "locksmithd.service"

// WriteLocksmithEnvironment writes a drop-in unit for locksmith
func WriteLocksmithEnvironment(strategy string, root string) error {
	cmd := "restart"
	if strategy == "off" {
		err := system.MaskUnit(locksmithUnit, root)
		if err != nil {
			return err
		}
		cmd = "stop"
	} else {
		unit := fmt.Sprintf("[Service]\nEnvironment=LOCKSMITH_STRATEGY=%s", strategy)
		file := system.File{
			Path:               path.Join(root, "run", "systemd", "system", locksmithUnit+".d", "20-cloudinit.conf"),
			RawFilePermissions: "0644",
			Content:            unit,
		}
		if err := system.WriteFile(&file); err != nil {
			return err
		}
	}
	if err := system.DaemonReload(); err != nil {
		return err
	}
	if _, err := system.RunUnitCommand(cmd, locksmithUnit); err != nil {
		return err
	}
	return nil
}
