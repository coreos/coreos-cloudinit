package initialize

import (
	"strings"
	"testing"
)

// Assert that the parsing of a cloud config file "generally works"
func TestCloudConfigEmpty(t *testing.T) {
	cfg, err := NewCloudConfig("")
	if err != nil {
		t.Fatalf("Encountered unexpected error :%v", err)
	}

	keys := cfg.SSHAuthorizedKeys
	if len(keys) != 0 {
		t.Error("Parsed incorrect number of SSH keys")
	}

	if len(cfg.WriteFiles) != 0 {
		t.Error("Expected zero WriteFiles")
	}

	if cfg.Hostname != "" {
		t.Errorf("Expected hostname to be empty, got '%s'", cfg.Hostname)
	}
}

// Assert that the parsing of a cloud config file "generally works"
func TestCloudConfig(t *testing.T) {
	contents := `
coreos: 
  etcd:
    discovery: "https://discovery.etcd.io/827c73219eeb2fa5530027c37bf18877"
  update:
    reboot-strategy: reboot
  units:
    - name: 50-eth0.network
      runtime: yes
      content: '[Match]
 
    Name=eth47
 
 
    [Network]
 
    Address=10.209.171.177/19
 
'
  oem:
    id: rackspace
    name: Rackspace Cloud Servers
    version-id: 168.0.0
    home-url: https://www.rackspace.com/cloud/servers/
    bug-report-url: https://github.com/coreos/coreos-overlay
ssh_authorized_keys:
  - foobar
  - foobaz
write_files:
  - content: |
      penny
      elroy
    path: /etc/dogepack.conf
    permissions: '0644'
    owner: root:dogepack
hostname: trontastic
`
	cfg, err := NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error :%v", err)
	}

	keys := cfg.SSHAuthorizedKeys
	if len(keys) != 2 {
		t.Error("Parsed incorrect number of SSH keys")
	} else if keys[0] != "foobar" {
		t.Error("Expected first SSH key to be 'foobar'")
	} else if keys[1] != "foobaz" {
		t.Error("Expected first SSH key to be 'foobaz'")
	}

	if len(cfg.WriteFiles) != 1 {
		t.Error("Failed to parse correct number of write_files")
	} else {
		wf := cfg.WriteFiles[0]
		if wf.Content != "penny\nelroy\n" {
			t.Errorf("WriteFile has incorrect contents '%s'", wf.Content)
		}
		if wf.Encoding != "" {
			t.Errorf("WriteFile has incorrect encoding %s", wf.Encoding)
		}
		if perm, _ := wf.Permissions(); perm != 0644 {
			t.Errorf("WriteFile has incorrect permissions %s", perm)
		}
		if wf.Path != "/etc/dogepack.conf" {
			t.Errorf("WriteFile has incorrect path %s", wf.Path)
		}
		if wf.Owner != "root:dogepack" {
			t.Errorf("WriteFile has incorrect owner %s", wf.Owner)
		}
	}

	if len(cfg.Coreos.Units) != 1 {
		t.Error("Failed to parse correct number of units")
	} else {
		u := cfg.Coreos.Units[0]
		expect := `[Match]
Name=eth47

[Network]
Address=10.209.171.177/19
`
		if u.Content != expect {
			t.Errorf("Unit has incorrect contents '%s'.\nExpected '%s'.", u.Content, expect)
		}
		if u.Runtime != true {
			t.Errorf("Unit has incorrect runtime value")
		}
		if u.Name != "50-eth0.network" {
			t.Errorf("Unit has incorrect name %s", u.Name)
		}
		if u.Type() != "network" {
			t.Errorf("Unit has incorrect type '%s'", u.Type())
		}
	}

	if cfg.Coreos.OEM.ID != "rackspace" {
		t.Errorf("Failed parsing coreos.oem. Expected ID 'rackspace', got %q.", cfg.Coreos.OEM.ID)
	}

	if cfg.Hostname != "trontastic" {
		t.Errorf("Failed to parse hostname")
	}
	if cfg.Coreos.Update["reboot-strategy"] != "reboot" {
		t.Errorf("Failed to parse locksmith strategy")
	}
}

// Assert that our interface conversion doesn't panic
func TestCloudConfigKeysNotList(t *testing.T) {
	contents := `
ssh_authorized_keys:
  - foo: bar
`
	cfg, err := NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error: %v", err)
	}

	keys := cfg.SSHAuthorizedKeys
	if len(keys) != 0 {
		t.Error("Parsed incorrect number of SSH keys")
	}
}

func TestCloudConfigSerializationHeader(t *testing.T) {
	cfg, _ := NewCloudConfig("")
	contents := cfg.String()
	header := strings.SplitN(contents, "\n", 2)[0]
	if header != "#cloud-config" {
		t.Fatalf("Serialized config did not have expected header")
	}
}

// TestDropInIgnored asserts that users are unable to set DropIn=True on units
func TestDropInIgnored(t *testing.T) {
	contents := `
coreos:
  units:
    - name: test
      dropin: true
`
	cfg, err := NewCloudConfig(contents)
	if err != nil || len(cfg.Coreos.Units) != 1 {
		t.Fatalf("Encountered unexpected error: %v", err)
	}
	if len(cfg.Coreos.Units) != 1 || cfg.Coreos.Units[0].Name != "test" {
		t.Fatalf("Expected 1 unit, but got %d: %v", len(cfg.Coreos.Units), cfg.Coreos.Units)
	}
	if cfg.Coreos.Units[0].DropIn {
		t.Errorf("dropin option on unit in cloud-config was not ignored!")
	}
}

func TestCloudConfigUsers(t *testing.T) {
	contents := `
users:
  - name: elroy
    passwd: somehash
    ssh-authorized-keys:
      - somekey
    gecos: arbitrary comment
    homedir: /home/place
    no-create-home: yes
    primary-group: things
    groups:
      - ping
      - pong
    no-user-group: true
    system: y
    no-log-init: True
`
	cfg, err := NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error: %v", err)
	}

	if len(cfg.Users) != 1 {
		t.Fatalf("Parsed %d users, expected 1", cfg.Users)
	}

	user := cfg.Users[0]

	if user.Name != "elroy" {
		t.Errorf("User name is %q, expected 'elroy'", user.Name)
	}

	if user.PasswordHash != "somehash" {
		t.Errorf("User passwd is %q, expected 'somehash'", user.PasswordHash)
	}

	if keys := user.SSHAuthorizedKeys; len(keys) != 1 {
		t.Errorf("Parsed %d ssh keys, expected 1", len(keys))
	} else {
		key := user.SSHAuthorizedKeys[0]
		if key != "somekey" {
			t.Errorf("User SSH key is %q, expected 'somekey'", key)
		}
	}

	if user.GECOS != "arbitrary comment" {
		t.Errorf("Failed to parse gecos field, got %q", user.GECOS)
	}

	if user.Homedir != "/home/place" {
		t.Errorf("Failed to parse homedir field, got %q", user.Homedir)
	}

	if !user.NoCreateHome {
		t.Errorf("Failed to parse no-create-home field")
	}

	if user.PrimaryGroup != "things" {
		t.Errorf("Failed to parse primary-group field, got %q", user.PrimaryGroup)
	}

	if len(user.Groups) != 2 {
		t.Errorf("Failed to parse 2 goups, got %d", len(user.Groups))
	} else {
		if user.Groups[0] != "ping" {
			t.Errorf("First group was %q, not expected value 'ping'", user.Groups[0])
		}
		if user.Groups[1] != "pong" {
			t.Errorf("First group was %q, not expected value 'pong'", user.Groups[1])
		}
	}

	if !user.NoUserGroup {
		t.Errorf("Failed to parse no-user-group field")
	}

	if !user.System {
		t.Errorf("Failed to parse system field")
	}

	if !user.NoLogInit {
		t.Errorf("Failed to parse no-log-init field")
	}
}
