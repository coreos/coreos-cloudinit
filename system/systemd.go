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

type Unit struct {
	Name    string
	Enable  bool
	Runtime bool
	Content string
	Command string
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

func PlaceUnit(u *Unit, root string) (string, error) {
	dir := "etc"
	if u.Runtime {
		dir = "run"
	}

	dst := path.Join(root, dir, "systemd", u.Group())
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		if err := os.MkdirAll(dst, os.FileMode(0755)); err != nil {
			return "", err
		}
	}

	dst = path.Join(dst, u.Name)

	file := File{
		Path: dst,
		Content: u.Content,
		RawFilePermissions: "0644",
	}

	err := WriteFile(&file)
	if err != nil {
		return "", err
	}

	return dst, nil
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
