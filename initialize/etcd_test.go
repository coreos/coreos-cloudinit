package initialize

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"
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

	if err := WriteEtcdEnvironment(ec, dir); err != nil {
		t.Fatalf("Processing of EtcdEnvironment failed: %v", err)
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
	ec := EtcdEnvironment{}
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

	if err := WriteEtcdEnvironment(ec, dir); err != nil {
		t.Fatalf("Processing of EtcdEnvironment failed: %v", err)
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

func rmdir(path string) error {
	cmd := exec.Command("rm", "-rf", path)
	return cmd.Run()
}
