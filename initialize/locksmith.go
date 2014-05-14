package initialize

import (
	"bufio"
	"os"
	"path"
	"strings"

	"github.com/coreos/coreos-cloudinit/system"
)

const locksmithUnit = "locksmithd.service"

type UpdateConfig map[string]string

func (uc UpdateConfig) strategy() string {
	s, _ := uc["reboot-strategy"]
	return s
}

// File creates an `/etc/coreos/update.conf` file with the requested
// strategy, by either rewriting the existing file on disk, or starting
// from `/usr/share/coreos/update.conf`
func (uc UpdateConfig) File(root string) (*system.File, error) {

	// If no reboot-strategy is set, we don't need to generate a new config
	if _, ok := uc["reboot-strategy"]; !ok {
		return nil, nil
	}

	var out string

	etcUpdate := path.Join(root, "etc", "coreos", "update.conf")
	usrUpdate := path.Join(root, "usr", "share", "coreos", "update.conf")

	conf, err := os.Open(etcUpdate)
	if os.IsNotExist(err) {
		conf, err = os.Open(usrUpdate)
	}
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(conf)

	sawStrat := false
	stratLine := "REBOOT_STRATEGY=" + uc.strategy()
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "REBOOT_STRATEGY=") {
			line = stratLine
			sawStrat = true
		}
		out += line
		out += "\n"
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	if !sawStrat {
		out += stratLine
		out += "\n"
	}
	return &system.File{
		Path:               path.Join("etc", "coreos", "update.conf"),
		RawFilePermissions: "0644",
		Content:            out,
	}, nil
}

// Unit generates a locksmith system.Unit for the cloud-init initializer to
// act on appropriately
func (uc UpdateConfig) Unit(root string) (*system.Unit, error) {
	u := &system.Unit{
		Name:    locksmithUnit,
		Enable:  true,
		Command: "restart",
		Mask:    false,
	}

	if uc.strategy() == "off" {
		u.Enable = false
		u.Command = "stop"
		u.Mask = true
	}

	return u, nil
}
