package initialize

import "testing"

func TestFleetEnvironment(t *testing.T) {
	cfg := make(FleetEnvironment, 0)
	cfg["public-ip"] = "12.34.56.78"

	env := cfg.String()

	expect := `[Service]
Environment="FLEET_PUBLIC_IP=12.34.56.78"
`

	if env != expect {
		t.Errorf("Generated environment:\n%s\nExpected environment:\n%s", env, expect)
	}
}

func TestFleetUnit(t *testing.T) {
	cfg := make(FleetEnvironment, 0)
	u, err := cfg.Unit("/")
	if u != nil {
		t.Errorf("unexpectedly generated unit with empty FleetEnvironment")
	}

	cfg["public-ip"] = "12.34.56.78"

	u, err = cfg.Unit("/")
	if err != nil {
		t.Errorf("error generating fleet unit: %v", err)
	}
	if u == nil {
		t.Fatalf("unexpectedly got nil unit generating fleet unit!")
	}
	if !u.Runtime {
		t.Errorf("bad Runtime for generated fleet unit!")
	}
	if !u.DropIn {
		t.Errorf("bad DropIn for generated fleet unit!")
	}
}
