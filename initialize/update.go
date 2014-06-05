package initialize

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/coreos/coreos-cloudinit/system"
)

const (
	locksmithUnit = "locksmithd.service"
)

// updateOption represents a configurable update option, which, if set, will be
// written into update.conf, replacing any existing value for the option
type updateOption struct {
	key    string   // key used to configure this option in cloud-config
	valid  []string // valid values for the option
	prefix string   // prefix for the option in the update.conf file
	value  string   // used to store the new value in update.conf (including prefix)
	seen   bool     // whether the option has been seen in any existing update.conf
}

// updateOptions defines the update options understood by cloud-config.
// The keys represent the string used in cloud-config to configure the option.
var updateOptions = []*updateOption{
	&updateOption{
		key:    "reboot-strategy",
		prefix: "REBOOT_STRATEGY=",
		valid:  []string{"best-effort", "etcd-lock", "reboot", "off"},
	},
	&updateOption{
		key:    "group",
		prefix: "GROUP=",
		valid:  []string{"master", "beta", "alpha", "stable"},
	},
	&updateOption{
		key:    "server",
		prefix: "SERVER=",
	},
}

// isValid checks whether a supplied value is valid for this option
func (uo updateOption) isValid(val string) bool {
	if len(uo.valid) == 0 {
		return true
	}
	for _, v := range uo.valid {
		if val == v {
			return true
		}
	}
	return false
}

type UpdateConfig map[string]string

// File generates an `/etc/coreos/update.conf` file (if any update
// configuration options are set in cloud-config) by either rewriting the
// existing file on disk, or starting from `/usr/share/coreos/update.conf`
func (uc UpdateConfig) File(root string) (*system.File, error) {
	if len(uc) < 1 {
		return nil, nil
	}

	var out string

	// Generate the list of possible substitutions to be performed based on the options that are configured
	subs := make([]*updateOption, 0)
	for _, uo := range updateOptions {
		val, ok := uc[uo.key]
		if !ok {
			continue
		}
		if !uo.isValid(val) {
			return nil, errors.New(fmt.Sprintf("invalid value %v for option %v (valid options: %v)", val, uo.key, uo.valid))
		}
		uo.value = uo.prefix + val
		subs = append(subs, uo)
	}

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

	for scanner.Scan() {
		line := scanner.Text()
		for _, s := range subs {
			if strings.HasPrefix(line, s.prefix) {
				line = s.value
				s.seen = true
				break
			}
		}
		out += line
		out += "\n"
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	for _, s := range subs {
		if !s.seen {
			out += s.value
			out += "\n"
		}
	}

	return &system.File{
		Path:               path.Join("etc", "coreos", "update.conf"),
		RawFilePermissions: "0644",
		Content:            out,
	}, nil
}

// Units generates units for the cloud-init initializer to act on:
// - a locksmith system.Unit, if "reboot-strategy" was set in cloud-config
// - an update_engine system.Unit, if "group" was set in cloud-config
func (uc UpdateConfig) Units(root string) ([]system.Unit, error) {
	var units []system.Unit
	if strategy, ok := uc["reboot-strategy"]; ok {
		ls := &system.Unit{
			Name:    locksmithUnit,
			Command: "restart",
			Mask:    false,
		}

		if strategy == "off" {
			ls.Command = "stop"
			ls.Mask = true
		}
		units = append(units, *ls)
	}

	rue := false
	if _, ok := uc["group"]; ok {
		rue = true
	}
	if _, ok := uc["server"]; ok {
		rue = true
	}
	if rue {
		ue := system.Unit{
			Name:    "update-engine",
			Command: "restart",
		}
		units = append(units, ue)
	}

	return units, nil
}
