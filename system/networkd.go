package system

import (
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"path"

	"github.com/coreos/coreos-cloudinit/network"
	"github.com/coreos/coreos-cloudinit/third_party/github.com/dotcloud/docker/pkg/netlink"
)

const (
	runtimeNetworkPath = "/run/systemd/network"
)

func RestartNetwork(interfaces []network.InterfaceGenerator) (err error) {
	defer func() {
		if e := restartNetworkd(); e != nil {
			err = e
		}
	}()

	if err = downNetworkInterfaces(interfaces); err != nil {
		return
	}

	if err = probe8012q(); err != nil {
		return
	}
	return
}

func downNetworkInterfaces(interfaces []network.InterfaceGenerator) error {
	sysInterfaceMap := make(map[string]*net.Interface)
	if systemInterfaces, err := net.Interfaces(); err == nil {
		for _, iface := range systemInterfaces {
			sysInterfaceMap[iface.Name] = &iface
		}
	} else {
		return err
	}

	for _, iface := range interfaces {
		if systemInterface, ok := sysInterfaceMap[iface.Name()]; ok {
			if err := netlink.NetworkLinkDown(systemInterface); err != nil {
				fmt.Printf("Error while downing interface %q (%s). Continuing...\n", systemInterface.Name, err)
			}
		}
	}

	return nil
}

func probe8012q() error {
	return exec.Command("modprobe", "8021q").Run()
}

func restartNetworkd() error {
	_, err := RunUnitCommand("restart", "systemd-networkd.service")
	return err
}

func WriteNetworkdConfigs(interfaces []network.InterfaceGenerator) error {
	for _, iface := range interfaces {
		filename := path.Join(runtimeNetworkPath, fmt.Sprintf("%s.netdev", iface.Name()))
		if err := writeConfig(filename, iface.Netdev()); err != nil {
			return err
		}
		filename = path.Join(runtimeNetworkPath, fmt.Sprintf("%s.link", iface.Name()))
		if err := writeConfig(filename, iface.Link()); err != nil {
			return err
		}
		filename = path.Join(runtimeNetworkPath, fmt.Sprintf("%s.network", iface.Name()))
		if err := writeConfig(filename, iface.Network()); err != nil {
			return err
		}
	}
	return nil
}

func writeConfig(filename string, config string) error {
	if config == "" {
		return nil
	}

	return ioutil.WriteFile(filename, []byte(config), 0444)
}
