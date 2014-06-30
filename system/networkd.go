package system

import (
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"path"
	"strings"

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

	if err = maybeProbe8012q(interfaces); err != nil {
		return
	}
	return maybeProbeBonding(interfaces)
}

func downNetworkInterfaces(interfaces []network.InterfaceGenerator) error {
	sysInterfaceMap := make(map[string]*net.Interface)
	if systemInterfaces, err := net.Interfaces(); err == nil {
		for _, iface := range systemInterfaces {
			// Need a copy of the interface so we can take the address
			temp := iface
			sysInterfaceMap[temp.Name] = &temp
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

func maybeProbe8012q(interfaces []network.InterfaceGenerator) error {
	for _, iface := range interfaces {
		if iface.Type() == "vlan" {
			return exec.Command("modprobe", "8021q").Run()
		}
	}
	return nil
}

func maybeProbeBonding(interfaces []network.InterfaceGenerator) error {
	args := []string{"bonding"}
	for _, iface := range interfaces {
		if iface.Type() == "bond" {
			args = append(args, strings.Split(iface.ModprobeParams(), " ")...)
			break
		}
	}
	return exec.Command("modprobe", args...).Run()
}

func restartNetworkd() error {
	_, err := RunUnitCommand("restart", "systemd-networkd.service")
	return err
}

func WriteNetworkdConfigs(interfaces []network.InterfaceGenerator) error {
	for _, iface := range interfaces {
		filename := path.Join(runtimeNetworkPath, fmt.Sprintf("%s.netdev", iface.Filename()))
		if err := writeConfig(filename, iface.Netdev()); err != nil {
			return err
		}
		filename = path.Join(runtimeNetworkPath, fmt.Sprintf("%s.link", iface.Filename()))
		if err := writeConfig(filename, iface.Link()); err != nil {
			return err
		}
		filename = path.Join(runtimeNetworkPath, fmt.Sprintf("%s.network", iface.Filename()))
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
