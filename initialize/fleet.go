package initialize

import (
	"fmt"

	"github.com/coreos/coreos-cloudinit/system"
)

type FleetEnvironment map[string]string

func (fe FleetEnvironment) String() (out string) {
	norm := normalizeSvcEnv(fe)
	out += "[Service]\n"

	for key, val := range norm {
		out += fmt.Sprintf("Environment=\"FLEET_%s=%s\"\n", key, val)
	}

	return
}

// Unit generates a Unit file drop-in for fleet, if any fleet options were
// configured in cloud-config
func (fe FleetEnvironment) Unit(root string) (*system.Unit, error) {
	if len(fe) < 1 {
		return nil, nil
	}
	return &system.Unit{
		Name:    "fleet.service",
		Runtime: true,
		DropIn:  true,
		Content: fe.String(),
	}, nil
}
