package network

import (
	"fmt"
	"strconv"
	"strings"
)

type InterfaceGenerator interface {
	Name() string
	Filename() string
	Netdev() string
	Link() string
	Network() string
	Type() string
	ModprobeParams() string
}

type networkInterface interface {
	InterfaceGenerator
	Children() []networkInterface
	setConfigDepth(int)
}

type logicalInterface struct {
	name        string
	config      configMethod
	children    []networkInterface
	configDepth int
}

func (i *logicalInterface) Network() string {
	config := fmt.Sprintf("[Match]\nName=%s\n\n[Network]\n", i.name)

	for _, child := range i.children {
		switch iface := child.(type) {
		case *vlanInterface:
			config += fmt.Sprintf("VLAN=%s\n", iface.name)
		case *bondInterface:
			config += fmt.Sprintf("Bond=%s\n", iface.name)
		}
	}

	switch conf := i.config.(type) {
	case configMethodStatic:
		for _, nameserver := range conf.nameservers {
			config += fmt.Sprintf("DNS=%s\n", nameserver)
		}
		if conf.address.IP != nil {
			config += fmt.Sprintf("\n[Address]\nAddress=%s\n", conf.address.String())
		}
		for _, route := range conf.routes {
			config += fmt.Sprintf("\n[Route]\nDestination=%s\nGateway=%s\n", route.destination.String(), route.gateway)
		}
	case configMethodDHCP:
		config += "DHCP=true\n"
	}

	return config
}

func (i *logicalInterface) Link() string {
	return ""
}

func (i *logicalInterface) Filename() string {
	return fmt.Sprintf("%02x-%s", i.configDepth, i.name)
}

func (i *logicalInterface) Children() []networkInterface {
	return i.children
}

func (i *logicalInterface) ModprobeParams() string {
	return ""
}

func (i *logicalInterface) setConfigDepth(depth int) {
	i.configDepth = depth
}

type physicalInterface struct {
	logicalInterface
}

func (p *physicalInterface) Name() string {
	return p.name
}

func (p *physicalInterface) Netdev() string {
	return ""
}

func (p *physicalInterface) Type() string {
	return "physical"
}

type bondInterface struct {
	logicalInterface
	slaves  []string
	options map[string]string
}

func (b *bondInterface) Name() string {
	return b.name
}

func (b *bondInterface) Netdev() string {
	return fmt.Sprintf("[NetDev]\nKind=bond\nName=%s\n", b.name)
}

func (b *bondInterface) Type() string {
	return "bond"
}

func (b *bondInterface) ModprobeParams() string {
	params := ""
	for name, val := range b.options {
		params += fmt.Sprintf("%s=%s ", name, val)
	}
	params = strings.TrimSuffix(params, " ")
	return params
}

type vlanInterface struct {
	logicalInterface
	id        int
	rawDevice string
}

func (v *vlanInterface) Name() string {
	return v.name
}

func (v *vlanInterface) Netdev() string {
	config := fmt.Sprintf("[NetDev]\nKind=vlan\nName=%s\n", v.name)
	switch c := v.config.(type) {
	case configMethodStatic:
		if c.hwaddress != nil {
			config += fmt.Sprintf("MACAddress=%s\n", c.hwaddress)
		}
	case configMethodDHCP:
		if c.hwaddress != nil {
			config += fmt.Sprintf("MACAddress=%s\n", c.hwaddress)
		}
	}
	config += fmt.Sprintf("\n[VLAN]\nId=%d\n", v.id)
	return config
}

func (v *vlanInterface) Type() string {
	return "vlan"
}

func buildInterfaces(stanzas []*stanzaInterface) []InterfaceGenerator {
	interfaceMap := createInterfaces(stanzas)
	linkAncestors(interfaceMap)
	markConfigDepths(interfaceMap)

	interfaces := make([]InterfaceGenerator, 0, len(interfaceMap))
	for _, iface := range interfaceMap {
		interfaces = append(interfaces, iface)
	}

	return interfaces
}

func createInterfaces(stanzas []*stanzaInterface) map[string]networkInterface {
	interfaceMap := make(map[string]networkInterface)
	for _, iface := range stanzas {
		switch iface.kind {
		case interfaceBond:
			bondOptions := make(map[string]string)
			for _, k := range []string{"mode", "miimon", "lacp-rate"} {
				if v, ok := iface.options["bond-"+k]; ok && len(v) > 0 {
					bondOptions[k] = v[0]
				}
			}
			interfaceMap[iface.name] = &bondInterface{
				logicalInterface{
					name:     iface.name,
					config:   iface.configMethod,
					children: []networkInterface{},
				},
				iface.options["bond-slaves"],
				bondOptions,
			}
			for _, slave := range iface.options["bond-slaves"] {
				if _, ok := interfaceMap[slave]; !ok {
					interfaceMap[slave] = &physicalInterface{
						logicalInterface{
							name:     slave,
							config:   configMethodManual{},
							children: []networkInterface{},
						},
					}
				}
			}

		case interfacePhysical:
			if _, ok := iface.configMethod.(configMethodLoopback); ok {
				continue
			}
			interfaceMap[iface.name] = &physicalInterface{
				logicalInterface{
					name:     iface.name,
					config:   iface.configMethod,
					children: []networkInterface{},
				},
			}

		case interfaceVLAN:
			var rawDevice string
			id, _ := strconv.Atoi(iface.options["id"][0])
			if device := iface.options["raw_device"]; len(device) == 1 {
				rawDevice = device[0]
				if _, ok := interfaceMap[rawDevice]; !ok {
					interfaceMap[rawDevice] = &physicalInterface{
						logicalInterface{
							name:     rawDevice,
							config:   configMethodManual{},
							children: []networkInterface{},
						},
					}
				}
			}
			interfaceMap[iface.name] = &vlanInterface{
				logicalInterface{
					name:     iface.name,
					config:   iface.configMethod,
					children: []networkInterface{},
				},
				id,
				rawDevice,
			}
		}
	}
	return interfaceMap
}

func linkAncestors(interfaceMap map[string]networkInterface) {
	for _, iface := range interfaceMap {
		switch i := iface.(type) {
		case *vlanInterface:
			if parent, ok := interfaceMap[i.rawDevice]; ok {
				switch p := parent.(type) {
				case *physicalInterface:
					p.children = append(p.children, iface)
				case *bondInterface:
					p.children = append(p.children, iface)
				}
			}
		case *bondInterface:
			for _, slave := range i.slaves {
				if parent, ok := interfaceMap[slave]; ok {
					switch p := parent.(type) {
					case *physicalInterface:
						p.children = append(p.children, iface)
					case *bondInterface:
						p.children = append(p.children, iface)
					}
				}
			}
		}
	}
}

func markConfigDepths(interfaceMap map[string]networkInterface) {
	rootInterfaceMap := make(map[string]networkInterface)
	for k, v := range interfaceMap {
		rootInterfaceMap[k] = v
	}

	for _, iface := range interfaceMap {
		for _, child := range iface.Children() {
			delete(rootInterfaceMap, child.Name())
		}
	}
	for _, iface := range rootInterfaceMap {
		setDepth(iface)
	}
}

func setDepth(iface networkInterface) int {
	maxDepth := 0
	for _, child := range iface.Children() {
		if depth := setDepth(child); depth > maxDepth {
			maxDepth = depth
		}
	}
	iface.setConfigDepth(maxDepth)
	return (maxDepth + 1)
}
