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
	bondStanzas := make(map[string]*stanzaInterface)
	physicalStanzas := make(map[string]*stanzaInterface)
	vlanStanzas := make(map[string]*stanzaInterface)
	for _, iface := range stanzas {
		switch iface.kind {
		case interfaceBond:
			bondStanzas[iface.name] = iface
		case interfacePhysical:
			physicalStanzas[iface.name] = iface
		case interfaceVLAN:
			vlanStanzas[iface.name] = iface
		}
	}

	physicals := make(map[string]*physicalInterface)
	for _, p := range physicalStanzas {
		if _, ok := p.configMethod.(configMethodLoopback); ok {
			continue
		}
		physicals[p.name] = &physicalInterface{
			logicalInterface{
				name:     p.name,
				config:   p.configMethod,
				children: []InterfaceGenerator{},
			},
		}
	}

	bonds := make(map[string]*bondInterface)
	for _, b := range bondStanzas {
		bonds[b.name] = &bondInterface{
			logicalInterface{
				name:     b.name,
				config:   b.configMethod,
				children: []InterfaceGenerator{},
			},
			b.options["slaves"],
		}
	}

	vlans := make(map[string]*vlanInterface)
	for _, v := range vlanStanzas {
		var rawDevice string
		id, _ := strconv.Atoi(v.options["id"][0])
		if device := v.options["raw_device"]; len(device) == 1 {
			rawDevice = device[0]
		}
		vlans[v.name] = &vlanInterface{
			logicalInterface{
				name:     v.name,
				config:   v.configMethod,
				children: []InterfaceGenerator{},
			},
			id,
			rawDevice,
		}
	}

	for _, vlan := range vlans {
		if physical, ok := physicals[vlan.rawDevice]; ok {
			physical.children = append(physical.children, vlan)
		}
		if bond, ok := bonds[vlan.rawDevice]; ok {
			bond.children = append(bond.children, vlan)
		}
	}

	for _, bond := range bonds {
		for _, slave := range bond.slaves {
			if physical, ok := physicals[slave]; ok {
				physical.children = append(physical.children, bond)
			}
			if pBond, ok := bonds[slave]; ok {
				pBond.children = append(pBond.children, bond)
			}
		}
	}

	interfaces := make([]InterfaceGenerator, 0, len(physicals)+len(bonds)+len(vlans))
	for _, physical := range physicals {
		interfaces = append(interfaces, physical)
	}
	for _, bond := range bonds {
		interfaces = append(interfaces, bond)
	}
	for _, vlan := range vlans {
		interfaces = append(interfaces, vlan)
	}

	return interfaces
}
