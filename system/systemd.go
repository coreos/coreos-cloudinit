package system

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/coreos/coreos-cloudinit/third_party/github.com/coreos/go-systemd/dbus"
)

// fakeMachineID is placed on non-usr CoreOS images and should
// never be used as a true MachineID
const fakeMachineID = "42000000000000000000000000000042"

// Name for drop-in service configuration files created by cloudconfig
const cloudConfigDropIn = "20-cloudinit.conf"

type Unit struct {
	Name    string
	Mask    bool
	Enable  bool
	Runtime bool
	Content string
	Command string

	// For drop-in units, a cloudinit.conf is generated.
	// This is currently unbound in YAML (and hence unsettable in cloud-config files)
	// until the correct behaviour for multiple drop-in units is determined.
	DropIn bool `yaml:"-"`
}

func (u *Unit) Type() string {
	ext := filepath.Ext(u.Name)
	return strings.TrimLeft(ext, ".")
}

func (u *Unit) Group() (group string) {
	t := u.Type()
	if t == "network" || t == "netdev" || t == "link" {
		group = "network"
	} else {
		group = "system"
	}
	return
}

type Script []byte

// UnitDestination builds the appropriate absolute file path for
// the given Unit. The root argument indicates the effective base
// directory of the system (similar to a chroot).
func UnitDestination(u *Unit, root string) string {
	dir := "etc"
	if u.Runtime {
		dir = "run"
	}

	if u.DropIn {
		return path.Join(root, dir, "systemd", u.Group(), fmt.Sprintf("%s.d", u.Name), cloudConfigDropIn)
	} else {
		return path.Join(root, dir, "systemd", u.Group(), u.Name)
	}
}

// PlaceUnit writes a unit file at the provided destination, creating
// parent directories as necessary.
func PlaceUnit(u *Unit, dst string) error {
	dir := filepath.Dir(dst)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.FileMode(0755)); err != nil {
			return err
		}
	}

	file := File{
		Path:               dst,
		Content:            u.Content,
		RawFilePermissions: "0644",
	}

	err := WriteFile(&file)
	if err != nil {
		return err
	}

	return nil
}

func EnableUnitFile(file string, runtime bool) error {
	conn, err := dbus.New()
	if err != nil {
		return err
	}

	files := []string{file}
	_, _, err = conn.EnableUnitFiles(files, runtime, true)
	return err
}

func RunUnitCommand(command, unit string) (string, error) {
	conn, err := dbus.New()
	if err != nil {
		return "", err
	}

	var fn func(string, string) (string, error)
	switch command {
	case "start":
		fn = conn.StartUnit
	case "stop":
		fn = conn.StopUnit
	case "restart":
		fn = conn.RestartUnit
	case "reload":
		fn = conn.ReloadUnit
	case "try-restart":
		fn = conn.TryRestartUnit
	case "reload-or-restart":
		fn = conn.ReloadOrRestartUnit
	case "reload-or-try-restart":
		fn = conn.ReloadOrTryRestartUnit
	default:
		return "", fmt.Errorf("Unsupported systemd command %q", command)
	}

	return fn(unit, "replace")
}

func DaemonReload() error {
	conn, err := dbus.New()
	if err != nil {
		return err
	}

	return conn.Reload()
}

func ExecuteScript(scriptPath string) (string, error) {
	props := []dbus.Property{
		dbus.PropDescription("Unit generated and executed by coreos-cloudinit on behalf of user"),
		dbus.PropExecStart([]string{"/bin/bash", scriptPath}, false),
	}

	base := path.Base(scriptPath)
	name := fmt.Sprintf("coreos-cloudinit-%s.service", base)

	log.Printf("Creating transient systemd unit '%s'", name)

	conn, err := dbus.New()
	if err != nil {
		return "", err
	}

	_, err = conn.StartTransientUnit(name, "replace", props...)
	return name, err
}

func SetHostname(hostname string) error {
	return exec.Command("hostnamectl", "set-hostname", hostname).Run()
}

func Hostname() (string, error) {
	return os.Hostname()
}

func MachineID(root string) string {
	contents, _ := ioutil.ReadFile(path.Join(root, "etc", "machine-id"))
	id := strings.TrimSpace(string(contents))

	if id == fakeMachineID {
		id = ""
	}

	return id
}

func MaskUnit(unit string, root string) error {
	masked := path.Join(root, "etc", "systemd", "system", unit)
	if err := os.MkdirAll(path.Dir(masked), os.FileMode(0755)); err != nil {
		return err
	}
	return os.Symlink("/dev/null", masked)
}
