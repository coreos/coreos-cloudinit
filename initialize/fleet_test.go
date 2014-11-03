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
	uu, err := cfg.Units("/")
	if len(uu) != 0 {
		t.Errorf("unexpectedly generated unit with empty FleetEnvironment")
	}

	cfg["public-ip"] = "12.34.56.78"

	uu, err = cfg.Units("/")
	if err != nil {
		t.Errorf("error generating fleet unit: %v", err)
	}
	if len(uu) != 1 {
		t.Fatalf("expected 1 unit generated, got %d", len(uu))
	}
	u := uu[0]
	if !u.Runtime {
		t.Errorf("bad Runtime for generated fleet unit!")
	}
	if !u.DropIn {
		t.Errorf("bad DropIn for generated fleet unit!")
	}
}
