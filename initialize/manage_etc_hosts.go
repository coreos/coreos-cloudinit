package initialize

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/coreos/coreos-cloudinit/system"
)

const DefaultIpv4Address = "127.0.0.1"

func generateEtcHosts(option string) (out string, err error) {
	if option != "localhost" {
		return "", errors.New("Invalid option to manage_etc_hosts")
	}

	// use the operating system hostname
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s\n", DefaultIpv4Address, hostname), nil

}

// Write an /etc/hosts file
func WriteEtcHosts(option string, root string) error {

	etcHosts, err := generateEtcHosts(option)
	if err != nil {
		return err
	}

	file := system.File{
		Path:               path.Join(root, "etc", "hosts"),
		RawFilePermissions: "0644",
		Content:            etcHosts,
	}

	return system.WriteFile(&file)
}
