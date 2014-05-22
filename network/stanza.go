package network

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type stanza interface{}

type stanzaAuto struct {
	interfaces []string
}

type stanzaInterface struct {
	name         string
	kind         interfaceKind
	auto         bool
	configMethod configMethod
	options      map[string][]string
}

type interfaceKind int

const (
	interfaceBond = interfaceKind(iota)
	interfacePhysical
	interfaceVLAN
)

type route struct {
	destination net.IPNet
	gateway     net.IP
}

type configMethod interface{}

type configMethodStatic struct {
	address     net.IPNet
	nameservers []net.IP
	routes      []route
}

type configMethodLoopback struct{}

type configMethodManual struct{}

type configMethodDHCP struct{}

func parseStanzas(lines []string) (stanzas []stanza, err error) {
	rawStanzas, err := splitStanzas(lines)
	if err != nil {
		return nil, err
	}

	stanzas = make([]stanza, 0, len(rawStanzas))
	for _, rawStanza := range rawStanzas {
		if stanza, err := parseStanza(rawStanza); err == nil {
			stanzas = append(stanzas, stanza)
		} else {
			return nil, err
		}
	}

	autos := make([]string, 0)
	interfaceMap := make(map[string]*stanzaInterface)
	for _, stanza := range stanzas {
		switch c := stanza.(type) {
		case *stanzaAuto:
			autos = append(autos, c.interfaces...)
		case *stanzaInterface:
			interfaceMap[c.name] = c
		}
	}

	// Apply the auto attribute
	for _, auto := range autos {
		if iface, ok := interfaceMap[auto]; ok {
			iface.auto = true
		}
	}

	return stanzas, nil
}

func splitStanzas(lines []string) ([][]string, error) {
	var curStanza []string
	stanzas := make([][]string, 0)
	for _, line := range lines {
		if isStanzaStart(line) {
			if curStanza != nil {
				stanzas = append(stanzas, curStanza)
			}
			curStanza = []string{line}
		} else if curStanza != nil {
			curStanza = append(curStanza, line)
		} else {
			return nil, fmt.Errorf("missing stanza start '%s'", line)
		}
	}

	if curStanza != nil {
		stanzas = append(stanzas, curStanza)
	}

	return stanzas, nil
}

func isStanzaStart(line string) bool {
	switch strings.Split(line, " ")[0] {
	case "auto":
		fallthrough
	case "iface":
		fallthrough
	case "mapping":
		return true
	}

	if strings.HasPrefix(line, "allow-") {
		return true
	}

	return false
}

func parseStanza(rawStanza []string) (stanza, error) {
	if len(rawStanza) == 0 {
		panic("empty stanza")
	}
	tokens := strings.Fields(rawStanza[0])
	if len(tokens) < 2 {
		return nil, fmt.Errorf("malformed stanza start %q", rawStanza[0])
	}

	kind := tokens[0]
	attributes := tokens[1:]

	switch kind {
	case "auto":
		return parseAutoStanza(attributes, rawStanza[1:])
	case "iface":
		return parseInterfaceStanza(attributes, rawStanza[1:])
	default:
		return nil, fmt.Errorf("unknown stanza '%s'", kind)
	}
}

func parseAutoStanza(attributes []string, options []string) (*stanzaAuto, error) {
	return &stanzaAuto{interfaces: attributes}, nil
}

func parseInterfaceStanza(attributes []string, options []string) (*stanzaInterface, error) {
	if len(attributes) != 3 {
		return nil, fmt.Errorf("incorrect number of attributes")
	}

	iface := attributes[0]
	confMethod := attributes[2]

	optionMap := make(map[string][]string, 0)
	for _, option := range options {
		if strings.HasPrefix(option, "post-up") {
			tokens := strings.SplitAfterN(option, " ", 2)
			if len(tokens) != 2 {
				continue
			}
			if v, ok := optionMap["post-up"]; ok {
				optionMap["post-up"] = append(v, tokens[1])
			} else {
				optionMap["post-up"] = []string{tokens[1]}
			}
		} else if strings.HasPrefix(option, "pre-down") {
			tokens := strings.SplitAfterN(option, " ", 2)
			if len(tokens) != 2 {
				continue
			}
			if v, ok := optionMap["pre-down"]; ok {
				optionMap["pre-down"] = append(v, tokens[1])
			} else {
				optionMap["pre-down"] = []string{tokens[1]}
			}
		} else {
			tokens := strings.Fields(option)
			optionMap[tokens[0]] = tokens[1:]
		}
	}

	var conf configMethod
	switch confMethod {
	case "static":
		config := configMethodStatic{
			routes:      make([]route, 0),
			nameservers: make([]net.IP, 0),
		}
		if addresses, ok := optionMap["address"]; ok {
			if len(addresses) == 1 {
				config.address.IP = net.ParseIP(addresses[0])
			}
		}
		if netmasks, ok := optionMap["netmask"]; ok {
			if len(netmasks) == 1 {
				config.address.Mask = net.IPMask(net.ParseIP(netmasks[0]).To4())
			}
		}
		if config.address.IP == nil || config.address.Mask == nil {
			return nil, fmt.Errorf("malformed static network config for '%s'", iface)
		}
		if gateways, ok := optionMap["gateway"]; ok {
			if len(gateways) == 1 {
				config.routes = append(config.routes, route{
					destination: net.IPNet{
						IP:   net.IPv4(0, 0, 0, 0),
						Mask: net.IPv4Mask(0, 0, 0, 0),
					},
					gateway: net.ParseIP(gateways[0]),
				})
			}
		}
		for _, nameserver := range optionMap["dns-nameservers"] {
			config.nameservers = append(config.nameservers, net.ParseIP(nameserver))
		}
		for _, postup := range optionMap["post-up"] {
			if strings.HasPrefix(postup, "route add") {
				route := route{}
				fields := strings.Fields(postup)
				for i, field := range fields[:len(fields)-1] {
					switch field {
					case "-net":
						route.destination.IP = net.ParseIP(fields[i+1])
					case "netmask":
						route.destination.Mask = net.IPMask(net.ParseIP(fields[i+1]).To4())
					case "gw":
						route.gateway = net.ParseIP(fields[i+1])
					}
				}
				if route.destination.IP != nil && route.destination.Mask != nil && route.gateway != nil {
					config.routes = append(config.routes, route)
				}
			}
		}
		conf = config
	case "loopback":
		conf = configMethodLoopback{}
	case "manual":
		conf = configMethodManual{}
	case "dhcp":
		conf = configMethodDHCP{}
	default:
		return nil, fmt.Errorf("invalid config method '%s'", confMethod)
	}

	if _, ok := optionMap["vlan_raw_device"]; ok {
		return parseVLANStanza(iface, conf, attributes, optionMap)
	}

	if strings.Contains(iface, ".") {
		return parseVLANStanza(iface, conf, attributes, optionMap)
	}

	if _, ok := optionMap["bond-slaves"]; ok {
		return parseBondStanza(iface, conf, attributes, optionMap)
	}

	return parsePhysicalStanza(iface, conf, attributes, optionMap)
}

func parseBondStanza(iface string, conf configMethod, attributes []string, options map[string][]string) (*stanzaInterface, error) {
	options["slaves"] = options["bond-slaves"]
	return &stanzaInterface{name: iface, kind: interfaceBond, configMethod: conf, options: options}, nil
}

func parsePhysicalStanza(iface string, conf configMethod, attributes []string, options map[string][]string) (*stanzaInterface, error) {
	return &stanzaInterface{name: iface, kind: interfacePhysical, configMethod: conf, options: options}, nil
}

func parseVLANStanza(iface string, conf configMethod, attributes []string, options map[string][]string) (*stanzaInterface, error) {
	var id string
	if strings.Contains(iface, ".") {
		tokens := strings.Split(iface, ".")
		id = tokens[len(tokens)-1]
	} else if strings.HasPrefix(iface, "vlan") {
		id = strings.TrimPrefix(iface, "vlan")
	} else {
		return nil, fmt.Errorf("malformed vlan name %s", iface)
	}

	if _, err := strconv.Atoi(id); err != nil {
		return nil, fmt.Errorf("malformed vlan name %s", iface)
	}
	options["id"] = []string{id}
	options["raw_device"] = options["vlan_raw_device"]

	return &stanzaInterface{name: iface, kind: interfaceVLAN, configMethod: conf, options: options}, nil
}
