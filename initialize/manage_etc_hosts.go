package initialize

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/coreos/coreos-cloudinit/system"
)

const DefaultIpv4Address = "127.0.0.1"

type EtcHosts string

func (eh EtcHosts) generateEtcHosts() (out string, err error) {
	if eh != "localhost" {
		return "", errors.New("Invalid option to manage_etc_hosts")
	}

	// use the operating system hostname
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s\n", DefaultIpv4Address, hostname), nil

}

func (eh EtcHosts) File(root string) (*system.File, error) {
	if eh == "" {
		return nil, nil
	}

	etcHosts, err := eh.generateEtcHosts()
	if err != nil {
		return nil, err
	}

	return &system.File{
		Path:               path.Join("etc", "hosts"),
		RawFilePermissions: "0644",
		Content:            etcHosts,
	}, nil
}
