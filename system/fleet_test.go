package system

import (
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func TestFleetUnits(t *testing.T) {
	for _, tt := range []struct {
		config config.Fleet
		units  []Unit
	}{
		{
			config.Fleet{},
			nil,
		},
		{
			config.Fleet{
				PublicIP: "12.34.56.78",
			},
			[]Unit{{
				Name: "fleet.service",
				Content: `[Service]
Environment="FLEET_PUBLIC_IP=12.34.56.78"
`,
				Runtime: true,
				DropIn:  true,
			}},
		},
	} {
		units, err := Fleet{tt.config}.Units("")
		if err != nil {
			t.Errorf("bad error (%q): want %q, got %q", tt.config, nil, err)
		}
		if !reflect.DeepEqual(units, tt.units) {
			t.Errorf("bad units (%q): want %q, got %q", tt.config, tt.units, units)
		}
	}
}
