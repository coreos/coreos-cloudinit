package network

import (
	"net"
	"reflect"
	"testing"
)

func TestPhysicalInterfaceName(t *testing.T) {
	p := physicalInterface{logicalInterface{name: "testname"}}
	if p.Name() != "testname" {
		t.FailNow()
	}
}

func TestPhysicalInterfaceNetdev(t *testing.T) {
	p := physicalInterface{}
	if p.Netdev() != "" {
		t.FailNow()
	}
}

func TestPhysicalInterfaceLink(t *testing.T) {
	p := physicalInterface{}
	if p.Link() != "" {
		t.FailNow()
	}
}

func TestPhysicalInterfaceNetwork(t *testing.T) {
	p := physicalInterface{logicalInterface{
		name: "testname",
		children: []networkInterface{
			&bondInterface{
				logicalInterface{
					name: "testbond1",
				},
				nil,
				nil,
			},
			&vlanInterface{
				logicalInterface{
					name: "testvlan1",
				},
				1,
				"",
			},
			&vlanInterface{
				logicalInterface{
					name: "testvlan2",
				},
				1,
				"",
			},
		},
	}}
	network := `[Match]
Name=testname

[Network]
Bond=testbond1
VLAN=testvlan1
VLAN=testvlan2
`
	if p.Network() != network {
		t.FailNow()
	}
}

func TestBondInterfaceName(t *testing.T) {
	b := bondInterface{logicalInterface{name: "testname"}, nil, nil}
	if b.Name() != "testname" {
		t.FailNow()
	}
}

func TestBondInterfaceNetdev(t *testing.T) {
	b := bondInterface{logicalInterface{name: "testname"}, nil, nil}
	netdev := `[NetDev]
Kind=bond
Name=testname
`
	if b.Netdev() != netdev {
		t.FailNow()
	}
}

func TestBondInterfaceLink(t *testing.T) {
	b := bondInterface{}
	if b.Link() != "" {
		t.FailNow()
	}
}

func TestBondInterfaceNetwork(t *testing.T) {
	b := bondInterface{
		logicalInterface{
			name:   "testname",
			config: configMethodDHCP{},
			children: []networkInterface{
				&bondInterface{
					logicalInterface{
						name: "testbond1",
					},
					nil,
					nil,
				},
				&vlanInterface{
					logicalInterface{
						name: "testvlan1",
					},
					1,
					"",
				},
				&vlanInterface{
					logicalInterface{
						name: "testvlan2",
					},
					1,
					"",
				},
			},
		},
		nil,
		nil,
	}
	network := `[Match]
Name=testname

[Network]
Bond=testbond1
VLAN=testvlan1
VLAN=testvlan2
DHCP=true
`
	if b.Network() != network {
		t.FailNow()
	}
}

func TestVLANInterfaceName(t *testing.T) {
	v := vlanInterface{logicalInterface{name: "testname"}, 1, ""}
	if v.Name() != "testname" {
		t.FailNow()
	}
}

func TestVLANInterfaceNetdev(t *testing.T) {
	for _, tt := range []struct {
		i vlanInterface
		l string
	}{
		{
			vlanInterface{logicalInterface{name: "testname"}, 1, ""},
			"[NetDev]\nKind=vlan\nName=testname\n\n[VLAN]\nId=1\n",
		},
		{
			vlanInterface{logicalInterface{name: "testname", config: configMethodStatic{hwaddress: net.HardwareAddr([]byte{0, 1, 2, 3, 4, 5})}}, 1, ""},
			"[NetDev]\nKind=vlan\nName=testname\nMACAddress=00:01:02:03:04:05\n\n[VLAN]\nId=1\n",
		},
		{
			vlanInterface{logicalInterface{name: "testname", config: configMethodDHCP{hwaddress: net.HardwareAddr([]byte{0, 1, 2, 3, 4, 5})}}, 1, ""},
			"[NetDev]\nKind=vlan\nName=testname\nMACAddress=00:01:02:03:04:05\n\n[VLAN]\nId=1\n",
		},
	} {
		if tt.i.Netdev() != tt.l {
			t.Fatalf("bad netdev config (%q): got %q, want %q", tt.i, tt.i.Netdev(), tt.l)
		}
	}
}

func TestVLANInterfaceLink(t *testing.T) {
	v := vlanInterface{}
	if v.Link() != "" {
		t.FailNow()
	}
}

func TestVLANInterfaceNetwork(t *testing.T) {
	v := vlanInterface{
		logicalInterface{
			name: "testname",
			config: configMethodStatic{
				address: net.IPNet{
					IP:   []byte{192, 168, 1, 100},
					Mask: []byte{255, 255, 255, 0},
				},
				nameservers: []net.IP{
					[]byte{8, 8, 8, 8},
				},
				routes: []route{
					route{
						destination: net.IPNet{
							IP:   []byte{0, 0, 0, 0},
							Mask: []byte{0, 0, 0, 0},
						},
						gateway: []byte{1, 2, 3, 4},
					},
				},
			},
		},
		0,
		"",
	}
	network := `[Match]
Name=testname

[Network]
DNS=8.8.8.8

[Address]
Address=192.168.1.100/24

[Route]
Destination=0.0.0.0/0
Gateway=1.2.3.4
`
	if v.Network() != network {
		t.Log(v.Network())
		t.FailNow()
	}
}

func TestType(t *testing.T) {
	for _, tt := range []struct {
		i InterfaceGenerator
		t string
	}{
		{
			i: &physicalInterface{},
			t: "physical",
		},
		{
			i: &vlanInterface{},
			t: "vlan",
		},
		{
			i: &bondInterface{},
			t: "bond",
		},
	} {
		if tp := tt.i.Type(); tp != tt.t {
			t.Fatalf("bad type (%q): got %s, want %s", tt.i, tp, tt.t)
		}
	}
}

func TestModprobeParams(t *testing.T) {
	for _, tt := range []struct {
		i InterfaceGenerator
		p string
	}{
		{
			i: &physicalInterface{},
			p: "",
		},
		{
			i: &vlanInterface{},
			p: "",
		},
		{
			i: &bondInterface{
				logicalInterface{},
				nil,
				map[string]string{
					"a": "1",
					"b": "2",
				},
			},
			p: "a=1 b=2",
		},
	} {
		if p := tt.i.ModprobeParams(); p != tt.p {
			t.Fatalf("bad params (%q): got %s, want %s", tt.i, p, tt.p)
		}
	}
}

func TestBuildInterfacesLo(t *testing.T) {
	stanzas := []*stanzaInterface{
		&stanzaInterface{
			name:         "lo",
			kind:         interfacePhysical,
			auto:         false,
			configMethod: configMethodLoopback{},
			options:      map[string][]string{},
		},
	}
	interfaces := buildInterfaces(stanzas)
	if len(interfaces) != 0 {
		t.FailNow()
	}
}

func TestBuildInterfacesBlindBond(t *testing.T) {
	stanzas := []*stanzaInterface{
		{
			name:         "bond0",
			kind:         interfaceBond,
			auto:         false,
			configMethod: configMethodManual{},
			options: map[string][]string{
				"bond-slaves": []string{"eth0"},
			},
		},
	}
	interfaces := buildInterfaces(stanzas)
	bond0 := &bondInterface{
		logicalInterface{
			name:        "bond0",
			config:      configMethodManual{},
			children:    []networkInterface{},
			configDepth: 0,
		},
		[]string{"eth0"},
		map[string]string{},
	}
	eth0 := &physicalInterface{
		logicalInterface{
			name:        "eth0",
			config:      configMethodManual{},
			children:    []networkInterface{bond0},
			configDepth: 1,
		},
	}
	expect := []InterfaceGenerator{bond0, eth0}
	if !reflect.DeepEqual(interfaces, expect) {
		t.FailNow()
	}
}

func TestBuildInterfacesBlindVLAN(t *testing.T) {
	stanzas := []*stanzaInterface{
		{
			name:         "vlan0",
			kind:         interfaceVLAN,
			auto:         false,
			configMethod: configMethodManual{},
			options: map[string][]string{
				"id":         []string{"0"},
				"raw_device": []string{"eth0"},
			},
		},
	}
	interfaces := buildInterfaces(stanzas)
	vlan0 := &vlanInterface{
		logicalInterface{
			name:        "vlan0",
			config:      configMethodManual{},
			children:    []networkInterface{},
			configDepth: 0,
		},
		0,
		"eth0",
	}
	eth0 := &physicalInterface{
		logicalInterface{
			name:        "eth0",
			config:      configMethodManual{},
			children:    []networkInterface{vlan0},
			configDepth: 1,
		},
	}
	expect := []InterfaceGenerator{eth0, vlan0}
	if !reflect.DeepEqual(interfaces, expect) {
		t.FailNow()
	}
}

func TestBuildInterfaces(t *testing.T) {
	stanzas := []*stanzaInterface{
		&stanzaInterface{
			name:         "eth0",
			kind:         interfacePhysical,
			auto:         false,
			configMethod: configMethodManual{},
			options:      map[string][]string{},
		},
		&stanzaInterface{
			name:         "bond0",
			kind:         interfaceBond,
			auto:         false,
			configMethod: configMethodManual{},
			options: map[string][]string{
				"bond-slaves": []string{"eth0"},
				"bond-mode":   []string{"4"},
				"bond-miimon": []string{"100"},
			},
		},
		&stanzaInterface{
			name:         "bond1",
			kind:         interfaceBond,
			auto:         false,
			configMethod: configMethodManual{},
			options: map[string][]string{
				"bond-slaves": []string{"bond0"},
			},
		},
		&stanzaInterface{
			name:         "vlan0",
			kind:         interfaceVLAN,
			auto:         false,
			configMethod: configMethodManual{},
			options: map[string][]string{
				"id":         []string{"0"},
				"raw_device": []string{"eth0"},
			},
		},
		&stanzaInterface{
			name:         "vlan1",
			kind:         interfaceVLAN,
			auto:         false,
			configMethod: configMethodManual{},
			options: map[string][]string{
				"id":         []string{"1"},
				"raw_device": []string{"bond0"},
			},
		},
	}
	interfaces := buildInterfaces(stanzas)
	vlan1 := &vlanInterface{
		logicalInterface{
			name:        "vlan1",
			config:      configMethodManual{},
			children:    []networkInterface{},
			configDepth: 0,
		},
		1,
		"bond0",
	}
	vlan0 := &vlanInterface{
		logicalInterface{
			name:        "vlan0",
			config:      configMethodManual{},
			children:    []networkInterface{},
			configDepth: 0,
		},
		0,
		"eth0",
	}
	bond1 := &bondInterface{
		logicalInterface{
			name:        "bond1",
			config:      configMethodManual{},
			children:    []networkInterface{},
			configDepth: 0,
		},
		[]string{"bond0"},
		map[string]string{},
	}
	bond0 := &bondInterface{
		logicalInterface{
			name:        "bond0",
			config:      configMethodManual{},
			children:    []networkInterface{bond1, vlan1},
			configDepth: 1,
		},
		[]string{"eth0"},
		map[string]string{
			"mode":   "4",
			"miimon": "100",
		},
	}
	eth0 := &physicalInterface{
		logicalInterface{
			name:        "eth0",
			config:      configMethodManual{},
			children:    []networkInterface{bond0, vlan0},
			configDepth: 2,
		},
	}
	expect := []InterfaceGenerator{eth0, bond0, bond1, vlan0, vlan1}
	if !reflect.DeepEqual(interfaces, expect) {
		t.FailNow()
	}
}

func TestFilename(t *testing.T) {
	for _, tt := range []struct {
		i logicalInterface
		f string
	}{
		{logicalInterface{name: "iface", configDepth: 0}, "00-iface"},
		{logicalInterface{name: "iface", configDepth: 9}, "09-iface"},
		{logicalInterface{name: "iface", configDepth: 10}, "0a-iface"},
		{logicalInterface{name: "iface", configDepth: 53}, "35-iface"},
	} {
		if tt.i.Filename() != tt.f {
			t.Fatalf("bad filename (%q): got %q, want %q", tt.i, tt.i.Filename(), tt.f)
		}
	}
}
