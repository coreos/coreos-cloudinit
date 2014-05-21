package initialize

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/coreos/coreos-cloudinit/system"
)

func TestEtcdEnvironment(t *testing.T) {
	cfg := make(EtcdEnvironment, 0)
	cfg["discovery"] = "http://disco.example.com/foobar"
	cfg["peer-bind-addr"] = "127.0.0.1:7002"

	env := cfg.String()
	expect := `[Service]
Environment="ETCD_DISCOVERY=http://disco.example.com/foobar"
Environment="ETCD_PEER_BIND_ADDR=127.0.0.1:7002"
`

	if env != expect {
		t.Errorf("Generated environment:\n%s\nExpected environment:\n%s", env, expect)
	}
}

func TestEtcdEnvironmentDiscoveryURLTranslated(t *testing.T) {
	cfg := make(EtcdEnvironment, 0)
	cfg["discovery_url"] = "http://disco.example.com/foobar"
	cfg["peer-bind-addr"] = "127.0.0.1:7002"

	env := cfg.String()
	expect := `[Service]
Environment="ETCD_DISCOVERY=http://disco.example.com/foobar"
Environment="ETCD_PEER_BIND_ADDR=127.0.0.1:7002"
`

	if env != expect {
		t.Errorf("Generated environment:\n%s\nExpected environment:\n%s", env, expect)
	}
}

func TestEtcdEnvironmentDiscoveryOverridesDiscoveryURL(t *testing.T) {
	cfg := make(EtcdEnvironment, 0)
	cfg["discovery_url"] = "ping"
	cfg["discovery"] = "pong"
	cfg["peer-bind-addr"] = "127.0.0.1:7002"

	env := cfg.String()
	expect := `[Service]
Environment="ETCD_DISCOVERY=pong"
Environment="ETCD_PEER_BIND_ADDR=127.0.0.1:7002"
`

	if env != expect {
		t.Errorf("Generated environment:\n%s\nExpected environment:\n%s", env, expect)
	}
}

func TestEtcdEnvironmentWrittenToDisk(t *testing.T) {
	ec := EtcdEnvironment{
		"name":           "node001",
		"discovery":      "http://disco.example.com/foobar",
		"peer-bind-addr": "127.0.0.1:7002",
	}
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	u, err := ec.Unit(dir)
	if err != nil {
		t.Fatalf("Generating etcd unit failed: %v", err)
	}
	if u == nil {
		t.Fatalf("Returned nil etcd unit unexpectedly")
	}

	dst := system.UnitDestination(u, dir)
	os.Stderr.WriteString("writing to " + dir + "\n")
	if err := system.PlaceUnit(u, dst); err != nil {
		t.Fatalf("Writing of EtcdEnvironment failed: %v", err)
	}

	fullPath := path.Join(dir, "run", "systemd", "system", "etcd.service.d", "20-cloudinit.conf")

	fi, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if fi.Mode() != os.FileMode(0644) {
		t.Errorf("File has incorrect mode: %v", fi.Mode())
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	expect := `[Service]
Environment="ETCD_NAME=node001"
Environment="ETCD_DISCOVERY=http://disco.example.com/foobar"
Environment="ETCD_PEER_BIND_ADDR=127.0.0.1:7002"
`
	if string(contents) != expect {
		t.Fatalf("File has incorrect contents")
	}
}

func TestEtcdEnvironmentWrittenToDiskDefaultToMachineID(t *testing.T) {
	ee := EtcdEnvironment{}
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	os.Mkdir(path.Join(dir, "etc"), os.FileMode(0755))
	err = ioutil.WriteFile(path.Join(dir, "etc", "machine-id"), []byte("node007"), os.FileMode(0444))
	if err != nil {
		t.Fatalf("Failed writing out /etc/machine-id: %v", err)
	}

	u, err := ee.Unit(dir)
	if err != nil {
		t.Fatalf("Generating etcd unit failed: %v", err)
	}
	if u == nil {
		t.Fatalf("Returned nil etcd unit unexpectedly")
	}

	dst := system.UnitDestination(u, dir)
	os.Stderr.WriteString("writing to " + dir + "\n")
	if err := system.PlaceUnit(u, dst); err != nil {
		t.Fatalf("Writing of EtcdEnvironment failed: %v", err)
	}

	fullPath := path.Join(dir, "run", "systemd", "system", "etcd.service.d", "20-cloudinit.conf")

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	expect := `[Service]
Environment="ETCD_NAME=node007"
`
	if string(contents) != expect {
		t.Fatalf("File has incorrect contents")
	}
}

func TestEtcdEnvironmentWhenNil(t *testing.T) {
	// EtcdEnvironment will be a nil map if it wasn't in the yaml
	var ee EtcdEnvironment
	if ee != nil {
		t.Fatalf("EtcdEnvironment is not nil")
	}
	u, err := ee.Unit("")
	if u != nil || err != nil {
		t.Fatalf("Unit returned a non-nil value for nil input")
	}
}
