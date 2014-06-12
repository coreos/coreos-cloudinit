package network

import (
	"fmt"
	"strconv"
)

type InterfaceGenerator interface {
	Name() string
	Netdev() string
	Link() string
	Network() string
}

type logicalInterface struct {
	name     string
	config   configMethod
	children []InterfaceGenerator
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

type physicalInterface struct {
	logicalInterface
}

func (p *physicalInterface) Name() string {
	return p.name
}

func (p *physicalInterface) Netdev() string {
	return ""
}

func (p *physicalInterface) Link() string {
	return ""
}

type bondInterface struct {
	logicalInterface
	slaves []string
}

func (b *bondInterface) Name() string {
	return b.name
}

func (b *bondInterface) Netdev() string {
	return fmt.Sprintf("[NetDev]\nKind=bond\nName=%s\n", b.name)
}

func (b *bondInterface) Link() string {
	return ""
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
	return fmt.Sprintf("[NetDev]\nKind=vlan\nName=%s\n\n[VLAN]\nId=%d\n", v.name, v.id)
}

func (v *vlanInterface) Link() string {
	return ""
}

func buildInterfaces(stanzas []*stanzaInterface) []InterfaceGenerator {
	interfaceMap := make(map[string]InterfaceGenerator)

	for _, iface := range stanzas {
		switch iface.kind {
		case interfaceBond:
			interfaceMap[iface.name] = &bondInterface{
				logicalInterface{
					name:     iface.name,
					config:   iface.configMethod,
					children: []InterfaceGenerator{},
				},
				iface.options["slaves"],
			}
			for _, slave := range iface.options["slaves"] {
				if _, ok := interfaceMap[slave]; !ok {
					interfaceMap[slave] = &physicalInterface{
						logicalInterface{
							name:     slave,
							config:   configMethodManual{},
							children: []InterfaceGenerator{},
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
					children: []InterfaceGenerator{},
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
							children: []InterfaceGenerator{},
						},
					}
				}
			}
			interfaceMap[iface.name] = &vlanInterface{
				logicalInterface{
					name:     iface.name,
					config:   iface.configMethod,
					children: []InterfaceGenerator{},
				},
				id,
				rawDevice,
			}
		}
	}

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

	interfaces := make([]InterfaceGenerator, 0, len(interfaceMap))
	for _, iface := range interfaceMap {
		interfaces = append(interfaces, iface)
	}

	return interfaces
}
