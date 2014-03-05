package cloudinit

import (
	"testing"
)

// Assert that the parsing of a cloud config file "generally works"
func TestCloudConfigEmpty(t *testing.T) {
	cfg, err := NewCloudConfig([]byte{})
	if err != nil {
		t.Fatalf("Encountered unexpected error :%v", err)
	}

	keys := cfg.SSH_Authorized_Keys
	if len(keys) != 0 {
		t.Error("Parsed incorrect number of SSH keys")
	}

	if cfg.Coreos.Etcd.Discovery_URL != "" {
		t.Error("Parsed incorrect value of discovery url")
	}

	if cfg.Coreos.Fleet.Autostart {
		t.Error("Expected AutostartFleet not to be defined")
	}
}

// Assert that the parsing of a cloud config file "generally works"
func TestCloudConfig(t *testing.T) {
	contents := []byte(`
coreos: 
  etcd:
    discovery_url: "https://discovery.etcd.io/827c73219eeb2fa5530027c37bf18877"
  fleet:
    autostart: Yes
ssh_authorized_keys:
  - foobar
  - foobaz
`)
	cfg, err := NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error :%v", err)
	}

	keys := cfg.SSH_Authorized_Keys
	if len(keys) != 2 {
		t.Error("Parsed incorrect number of SSH keys")
	} else if keys[0] != "foobar" {
		t.Error("Expected first SSH key to be 'foobar'")
	} else if keys[1] != "foobaz" {
		t.Error("Expected first SSH key to be 'foobaz'")
	}

	if cfg.Coreos.Etcd.Discovery_URL != "https://discovery.etcd.io/827c73219eeb2fa5530027c37bf18877" {
		t.Error("Failed to parse etcd discovery url")
	}

	if !cfg.Coreos.Fleet.Autostart {
		t.Error("Expected AutostartFleet to be true")
	}
}

// Assert that our interface conversion doesn't panic
func TestCloudConfigKeysNotList(t *testing.T) {
	contents := []byte(`
ssh_authorized_keys:
  - foo: bar
`)
	cfg, err := NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error :%v", err)
	}

	keys := cfg.SSH_Authorized_Keys
	if len(keys) != 0 {
		t.Error("Parsed incorrect number of SSH keys")
	}
}
