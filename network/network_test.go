package network

import (
	"testing"
)

func TestFormatConfigs(t *testing.T) {
	for in, n := range map[string]int{
		"":                                                    0,
		"line1\\\nis long":                                    1,
		"#comment":                                            0,
		"#comment\\\ncomment":                                 0,
		"  #comment \\\n comment\nline 1\nline 2\\\n is long": 2,
	} {
		lines := formatConfig(in)
		if len(lines) != n {
			t.Fatalf("bad number of lines for config %q: got %d, want %d", in, len(lines), n)
		}
	}
}

func TestProcessDebianNetconf(t *testing.T) {
	for _, tt := range []struct {
		in   string
		fail bool
		n    int
	}{
		{"", false, 0},
		{"iface", true, -1},
		{"auto eth1\nauto eth2", false, 0},
		{"iface eth1 inet manual", false, 1},
	} {
		interfaces, err := ProcessDebianNetconf(tt.in)
		failed := err != nil
		if tt.fail != failed {
			t.Fatalf("bad failure state for %q: got %b, want %b", failed, tt.fail)
		}
		if tt.n != -1 && tt.n != len(interfaces) {
			t.Fatalf("bad number of interfaces for %q: got %d, want %q", tt.in, len(interfaces), tt.n)
		}
	}
}
