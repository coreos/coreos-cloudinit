package system

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

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
			return
		}
		// TODO(crawford): Get rid of this once networkd fixes the race
		// https://bugs.freedesktop.org/show_bug.cgi?id=76077
		time.Sleep(5 * time.Second)
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
			iface := iface
			sysInterfaceMap[iface.Name] = &iface
		}
	} else {
		return err
	}

	for _, iface := range interfaces {
		if systemInterface, ok := sysInterfaceMap[iface.Name()]; ok {
			log.Printf("Taking down interface %q\n", systemInterface.Name)
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
			log.Printf("Probing LKM %q (%q)\n", "8021q", "8021q")
			return exec.Command("modprobe", "8021q").Run()
		}
	}
	return nil
}

func maybeProbeBonding(interfaces []network.InterfaceGenerator) error {
	for _, iface := range interfaces {
		if iface.Type() == "bond" {
			args := append([]string{"bonding"}, strings.Split(iface.ModprobeParams(), " ")...)
			log.Printf("Probing LKM %q (%q)\n", "bonding", args)
			return exec.Command("modprobe", args...).Run()
		}
	}
	return nil
}

func restartNetworkd() error {
	log.Printf("Restarting networkd.service\n")
	_, err := NewUnitManager("").RunUnitCommand("restart", "systemd-networkd.service")
	return err
}

func WriteNetworkdConfigs(interfaces []network.InterfaceGenerator) error {
	for _, iface := range interfaces {
		filename := fmt.Sprintf("%s.netdev", iface.Filename())
		if err := writeConfig(filename, iface.Netdev()); err != nil {
			return err
		}
		filename = fmt.Sprintf("%s.link", iface.Filename())
		if err := writeConfig(filename, iface.Link()); err != nil {
			return err
		}
		filename = fmt.Sprintf("%s.network", iface.Filename())
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
	log.Printf("Writing networkd unit %q\n", filename)
	_, err := WriteFile(&File{Content: config, Path: filename}, runtimeNetworkPath)
	return err
}
