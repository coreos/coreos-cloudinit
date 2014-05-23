package network

import (
	"net"
	"reflect"
	"strings"
	"testing"
)

func TestSplitStanzasNoParent(t *testing.T) {
	in := []string{"test"}
	e := "missing stanza start"
	_, err := splitStanzas(in)
	if err == nil || !strings.HasPrefix(err.Error(), e) {
		t.Fatalf("bad error for splitStanzas(%q): got %q, want %q", in, err, e)
	}
}

func TestBadParseStanzas(t *testing.T) {
	for in, e := range map[string]string{
		"":                 "missing stanza start",
		"iface":            "malformed stanza start",
		"allow-?? unknown": "unknown stanza",
	} {
		_, err := parseStanzas([]string{in})
		if err == nil || !strings.HasPrefix(err.Error(), e) {
			t.Fatalf("bad error for parseStanzas(%q): got %q, want %q", in, err, e)
		}

	}
}

func TestBadParseInterfaceStanza(t *testing.T) {
	for _, tt := range []struct {
		in   []string
		opts []string
		e    string
	}{
		{[]string{}, nil, "incorrect number of attributes"},
		{[]string{"eth", "inet", "invalid"}, nil, "invalid config method"},
		{[]string{"eth", "inet", "static"}, []string{"address 192.168.1.100"}, "malformed static network config"},
		{[]string{"eth", "inet", "static"}, []string{"netmask 255.255.255.0"}, "malformed static network config"},
		{[]string{"eth", "inet", "static"}, []string{"address invalid", "netmask 255.255.255.0"}, "malformed static network config"},
		{[]string{"eth", "inet", "static"}, []string{"address 192.168.1.100", "netmask invalid"}, "malformed static network config"},
	} {
		_, err := parseInterfaceStanza(tt.in, tt.opts)
		if err == nil || !strings.HasPrefix(err.Error(), tt.e) {
			t.Fatalf("bad error parsing interface stanza %q: got %q, want %q", tt.in, err.Error(), tt.e)
		}
	}
}

func TestBadParseVLANStanzas(t *testing.T) {
	conf := configMethodManual{}
	options := map[string][]string{}
	for _, in := range []string{"myvlan", "eth.vlan"} {
		_, err := parseVLANStanza(in, conf, nil, options)
		if err == nil || !strings.HasPrefix(err.Error(), "malformed vlan name") {
			t.Fatalf("did not error on bad vlan %q", in)
		}
	}
}

func TestSplitStanzas(t *testing.T) {
	expect := [][]string{
		{"auto lo"},
		{"iface eth1", "option: 1"},
		{"mapping"},
		{"allow-"},
	}
	lines := make([]string, 0, 5)
	for _, stanza := range expect {
		for _, line := range stanza {
			lines = append(lines, line)
		}
	}

	stanzas, err := splitStanzas(lines)
	if err != nil {
		t.FailNow()
	}
	for i, stanza := range stanzas {
		if len(stanza) != len(expect[i]) {
			t.FailNow()
		}
		for j, line := range stanza {
			if line != expect[i][j] {
				t.FailNow()
			}
		}
	}
}

func TestParseStanzaNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("parseStanza(nil) did not panic")
		}
	}()
	parseStanza(nil)
}

func TestParseStanzaSuccess(t *testing.T) {
	for _, in := range []string{
		"auto a",
		"iface a inet manual",
	} {
		if _, err := parseStanza([]string{in}); err != nil {
			t.Fatalf("unexpected error parsing stanza %q: %s", in, err)
		}
	}
}

func TestParseAutoStanza(t *testing.T) {
	interfaces := []string{"test", "attribute"}
	stanza, err := parseAutoStanza(interfaces, nil)
	if err != nil {
		t.Fatalf("unexpected error parsing auto stanza %q: %s", interfaces, err)
	}
	if !reflect.DeepEqual(stanza.interfaces, interfaces) {
		t.FailNow()
	}
}

func TestParseBondStanzaNoSlaves(t *testing.T) {
	bond, err := parseBondStanza("", nil, nil, map[string][]string{})
	if err != nil {
		t.FailNow()
	}
	if bond.options["slaves"] != nil {
		t.FailNow()
	}
}

func TestParseBondStanza(t *testing.T) {
	conf := configMethodManual{}
	options := map[string][]string{
		"bond-slaves": []string{"1", "2"},
	}
	bond, err := parseBondStanza("test", conf, nil, options)
	if err != nil {
		t.FailNow()
	}
	if bond.name != "test" {
		t.FailNow()
	}
	if bond.kind != interfaceBond {
		t.FailNow()
	}
	if bond.configMethod != conf {
		t.FailNow()
	}
	if !reflect.DeepEqual(bond.options["slaves"], options["bond-slaves"]) {
		t.FailNow()
	}
}

func TestParsePhysicalStanza(t *testing.T) {
	conf := configMethodManual{}
	options := map[string][]string{
		"a": []string{"1", "2"},
		"b": []string{"1"},
	}
	physical, err := parsePhysicalStanza("test", conf, nil, options)
	if err != nil {
		t.FailNow()
	}
	if physical.name != "test" {
		t.FailNow()
	}
	if physical.kind != interfacePhysical {
		t.FailNow()
	}
	if physical.configMethod != conf {
		t.FailNow()
	}
	if !reflect.DeepEqual(physical.options, options) {
		t.FailNow()
	}
}

func TestParseVLANStanzas(t *testing.T) {
	conf := configMethodManual{}
	options := map[string][]string{}
	for _, in := range []string{"vlan25", "eth.25"} {
		vlan, err := parseVLANStanza(in, conf, nil, options)
		if err != nil {
			t.Fatalf("unexpected error from parseVLANStanza(%q): %s", in, err)
		}
		if !reflect.DeepEqual(vlan.options["id"], []string{"25"}) {
			t.FailNow()
		}
	}
}

func TestParseInterfaceStanzaStaticAddress(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0"}
	expect := net.IPNet{
		IP:   net.IPv4(192, 168, 1, 100),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}

	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if !reflect.DeepEqual(static.address, expect) {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticGateway(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0", "gateway 192.168.1.1"}
	expect := []route{
		{
			destination: net.IPNet{
				IP:   net.IPv4(0, 0, 0, 0),
				Mask: net.IPv4Mask(0, 0, 0, 0),
			},
			gateway: net.IPv4(192, 168, 1, 1),
		},
	}

	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if !reflect.DeepEqual(static.routes, expect) {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaStaticDNS(t *testing.T) {
	options := []string{"address 192.168.1.100", "netmask 255.255.255.0", "dns-nameservers 192.168.1.10 192.168.1.11 192.168.1.12"}
	expect := []net.IP{
		net.IPv4(192, 168, 1, 10),
		net.IPv4(192, 168, 1, 11),
		net.IPv4(192, 168, 1, 12),
	}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if !reflect.DeepEqual(static.nameservers, expect) {
		t.FailNow()
	}
}

func TestBadParseInterfaceStanzasStaticPostUp(t *testing.T) {
	for _, in := range []string{
		"post-up invalid",
		"post-up route add",
		"post-up route add -net",
		"post-up route add gw",
		"post-up route add netmask",
		"gateway",
		"gateway 192.168.1.1 192.168.1.2",
	} {
		options := []string{"address 192.168.1.100", "netmask 255.255.255.0", in}
		iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
		if err != nil {
			t.Fatalf("parseInterfaceStanza with options %s got unexpected error", options)
		}
		static, ok := iface.configMethod.(configMethodStatic)
		if !ok {
			t.Fatalf("parseInterfaceStanza with options %s did not return configMethodStatic", options)
		}
		if len(static.routes) != 0 {
			t.Fatalf("parseInterfaceStanza with options %s did not return zero-length static routes", options)
		}
	}
}

func TestParseInterfaceStanzaStaticPostUp(t *testing.T) {
	options := []string{
		"address 192.168.1.100",
		"netmask 255.255.255.0",
		"post-up route add gw 192.168.1.1 -net 192.168.1.0 netmask 255.255.255.0",
	}
	expect := []route{
		{
			destination: net.IPNet{
				IP:   net.IPv4(192, 168, 1, 0),
				Mask: net.IPv4Mask(255, 255, 255, 0),
			},
			gateway: net.IPv4(192, 168, 1, 1),
		},
	}

	iface, err := parseInterfaceStanza([]string{"eth", "inet", "static"}, options)
	if err != nil {
		t.FailNow()
	}
	static, ok := iface.configMethod.(configMethodStatic)
	if !ok {
		t.FailNow()
	}
	if !reflect.DeepEqual(static.routes, expect) {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaLoopback(t *testing.T) {
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "loopback"}, nil)
	if err != nil {
		t.FailNow()
	}
	if _, ok := iface.configMethod.(configMethodLoopback); !ok {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaManual(t *testing.T) {
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "manual"}, nil)
	if err != nil {
		t.FailNow()
	}
	if _, ok := iface.configMethod.(configMethodManual); !ok {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaDHCP(t *testing.T) {
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "dhcp"}, nil)
	if err != nil {
		t.FailNow()
	}
	if _, ok := iface.configMethod.(configMethodDHCP); !ok {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaPostUpOption(t *testing.T) {
	options := []string{
		"post-up",
		"post-up 1 2",
		"post-up 3 4",
	}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "manual"}, options)
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(iface.options["post-up"], []string{"1 2", "3 4"}) {
		t.Log(iface.options["post-up"])
		t.FailNow()
	}
}

func TestParseInterfaceStanzaPreDownOption(t *testing.T) {
	options := []string{
		"pre-down",
		"pre-down 3",
		"pre-down 4",
	}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "manual"}, options)
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(iface.options["pre-down"], []string{"3", "4"}) {
		t.Log(iface.options["pre-down"])
		t.FailNow()
	}
}

func TestParseInterfaceStanzaEmptyOption(t *testing.T) {
	options := []string{
		"test",
	}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "manual"}, options)
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(iface.options["test"], []string{}) {
		t.FailNow()
	}
}

func TestParseInterfaceStanzaOptions(t *testing.T) {
	options := []string{
		"test1 1",
		"test2 2 3",
		"test1 5 6",
	}
	iface, err := parseInterfaceStanza([]string{"eth", "inet", "manual"}, options)
	if err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(iface.options["test1"], []string{"5", "6"}) {
		t.Log(iface.options["test1"])
		t.FailNow()
	}
	if !reflect.DeepEqual(iface.options["test2"], []string{"2", "3"}) {
		t.Log(iface.options["test2"])
		t.FailNow()
	}
}

func TestParseInterfaceStazaBond(t *testing.T) {
	iface, err := parseInterfaceStanza([]string{"mybond", "inet", "manual"}, []string{"bond-slaves eth"})
	if err != nil {
		t.FailNow()
	}
	if iface.kind != interfaceBond {
		t.FailNow()
	}
}

func TestParseInterfaceStazaVLANName(t *testing.T) {
	iface, err := parseInterfaceStanza([]string{"eth0.1", "inet", "manual"}, nil)
	if err != nil {
		t.FailNow()
	}
	if iface.kind != interfaceVLAN {
		t.FailNow()
	}
}

func TestParseInterfaceStazaVLANOption(t *testing.T) {
	iface, err := parseInterfaceStanza([]string{"vlan1", "inet", "manual"}, []string{"vlan_raw_device eth"})
	if err != nil {
		t.FailNow()
	}
	if iface.kind != interfaceVLAN {
		t.FailNow()
	}
}

func TestParseStanzasNone(t *testing.T) {
	stanzas, err := parseStanzas(nil)
	if err != nil {
		t.FailNow()
	}
	if len(stanzas) != 0 {
		t.FailNow()
	}
}

func TestParseStanzas(t *testing.T) {
	lines := []string{
		"auto lo",
		"iface lo inet loopback",
		"iface eth1 inet manual",
		"iface eth2 inet manual",
		"iface eth3 inet manual",
		"auto eth1 eth3",
	}
	expect := []stanza{
		&stanzaAuto{
			interfaces: []string{"lo"},
		},
		&stanzaInterface{
			name:         "lo",
			kind:         interfacePhysical,
			auto:         true,
			configMethod: configMethodLoopback{},
			options:      map[string][]string{},
		},
		&stanzaInterface{
			name:         "eth1",
			kind:         interfacePhysical,
			auto:         true,
			configMethod: configMethodManual{},
			options:      map[string][]string{},
		},
		&stanzaInterface{
			name:         "eth2",
			kind:         interfacePhysical,
			auto:         false,
			configMethod: configMethodManual{},
			options:      map[string][]string{},
		},
		&stanzaInterface{
			name:         "eth3",
			kind:         interfacePhysical,
			auto:         true,
			configMethod: configMethodManual{},
			options:      map[string][]string{},
		},
		&stanzaAuto{
			interfaces: []string{"eth1", "eth3"},
		},
	}
	stanzas, err := parseStanzas(lines)
	if err != err {
		t.FailNow()
	}
	if !reflect.DeepEqual(stanzas, expect) {
		t.FailNow()
	}
}
