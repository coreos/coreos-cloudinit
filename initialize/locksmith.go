package initialize

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/coreos/coreos-cloudinit/system"
)

const locksmithUnit = "locksmithd.service"

// addStrategy creates an `/etc/coreos/update.conf` file with the requested
// strategy via rewriting the file on disk or by starting from
// `/usr/share/coreos/update.conf`.
func addStrategy(strategy string, root string) error {
	etcUpdate := path.Join(root, "etc", "coreos", "update.conf")
	usrUpdate := path.Join(root, "usr", "share", "coreos", "update.conf")

	tmp, err := ioutil.TempFile(path.Join(root, "etc", "coreos"), ".update.conf")
	if err != nil {
		return err
	}
	err = tmp.Chmod(0644)
	if err != nil {
		return err
	}

	conf, err := os.Open(etcUpdate)
	if os.IsNotExist(err) {
		conf, err = os.Open(usrUpdate)
		if err != nil {
			return err
		}
	}

	scanner := bufio.NewScanner(conf)

	sawStrat := false
	stratLine := "STRATEGY="+strategy
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "STRATEGY=") {
			line = stratLine
			sawStrat = true
		}
		fmt.Fprintln(tmp, line)
		if err := scanner.Err(); err != nil {
			return err
		}
	}

	if !sawStrat {
		fmt.Fprintln(tmp, stratLine)
	}

	return os.Rename(tmp.Name(), etcUpdate)
}

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
		return addStrategy(strategy, root)
	}
	if err := system.DaemonReload(); err != nil {
		return err
	}
	if _, err := system.RunUnitCommand(cmd, locksmithUnit); err != nil {
		return err
	}
	return nil
}
