/*
   Copyright 2014 CoreOS, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package config

import (
	"reflect"
	"strings"
	"testing"
)

func TestIsZero(t *testing.T) {
	for _, tt := range []struct {
		c     interface{}
		empty bool
	}{
		{struct{}{}, true},
		{struct{ a, b string }{}, true},
		{struct{ A, b string }{}, true},
		{struct{ A, B string }{}, true},
		{struct{ A string }{A: "hello"}, false},
		{struct{ A int }{}, true},
		{struct{ A int }{A: 1}, false},
	} {
		if empty := IsZero(tt.c); tt.empty != empty {
			t.Errorf("bad result (%q): want %t, got %t", tt.c, tt.empty, empty)
		}
	}
}

func TestAssertStructValid(t *testing.T) {
	for _, tt := range []struct {
		c   interface{}
		err error
	}{
		{struct{}{}, nil},
		{struct {
			A, b string `valid:"1,2"`
		}{}, nil},
		{struct {
			A, b string `valid:"1,2"`
		}{A: "1", b: "2"}, nil},
		{struct {
			A, b string `valid:"1,2"`
		}{A: "1", b: "hello"}, nil},
		{struct {
			A, b string `valid:"1,2"`
		}{A: "hello", b: "2"}, &ErrorValid{Value: "hello", Field: "A", Valid: []string{"1", "2"}}},
		{struct {
			A, b int `valid:"1,2"`
		}{}, nil},
		{struct {
			A, b int `valid:"1,2"`
		}{A: 1, b: 2}, nil},
		{struct {
			A, b int `valid:"1,2"`
		}{A: 1, b: 9}, nil},
		{struct {
			A, b int `valid:"1,2"`
		}{A: 9, b: 2}, &ErrorValid{Value: "9", Field: "A", Valid: []string{"1", "2"}}},
	} {
		if err := AssertStructValid(tt.c); !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad result (%q): want %q, got %q", tt.c, tt.err, err)
		}
	}
}

func TestCloudConfigInvalidKeys(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic while instantiating CloudConfig with nil keys: %v", r)
		}
	}()

	for _, tt := range []struct {
		contents string
	}{
		{"coreos:"},
		{"ssh_authorized_keys:"},
		{"ssh_authorized_keys:\n  -"},
		{"ssh_authorized_keys:\n  - 0:"},
		{"write_files:"},
		{"write_files:\n  -"},
		{"write_files:\n  - 0:"},
		{"users:"},
		{"users:\n  -"},
		{"users:\n  - 0:"},
	} {
		_, err := NewCloudConfig(tt.contents)
		if err != nil {
			t.Fatalf("error instantiating CloudConfig with invalid keys: %v", err)
		}
	}
}

func TestCloudConfigUnknownKeys(t *testing.T) {
	contents := `
coreos: 
  etcd:
    discovery: "https://discovery.etcd.io/827c73219eeb2fa5530027c37bf18877"
  coreos_unknown:
    foo: "bar"
section_unknown:
  dunno:
    something
bare_unknown:
  bar
write_files:
  - content: fun
    path: /var/party
    file_unknown: nofun
users:
  - name: fry
    passwd: somehash
    user_unknown: philip
hostname:
  foo
`
	cfg, err := NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("error instantiating CloudConfig with unknown keys: %v", err)
	}
	if cfg.Hostname != "foo" {
		t.Fatalf("hostname not correctly set when invalid keys are present")
	}
	if cfg.Coreos.Etcd.Discovery != "https://discovery.etcd.io/827c73219eeb2fa5530027c37bf18877" {
		t.Fatalf("etcd section not correctly set when invalid keys are present")
	}
	if len(cfg.WriteFiles) < 1 || cfg.WriteFiles[0].Content != "fun" || cfg.WriteFiles[0].Path != "/var/party" {
		t.Fatalf("write_files section not correctly set when invalid keys are present")
	}
	if len(cfg.Users) < 1 || cfg.Users[0].Name != "fry" || cfg.Users[0].PasswordHash != "somehash" {
		t.Fatalf("users section not correctly set when invalid keys are present")
	}
}

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
    reboot_strategy: reboot
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
    version_id: 168.0.0
    home_url: https://www.rackspace.com/cloud/servers/
    bug_report_url: https://github.com/coreos/coreos-overlay
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
		if wf.RawFilePermissions != "0644" {
			t.Errorf("WriteFile has incorrect permissions %s", wf.RawFilePermissions)
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
	}

	if cfg.Coreos.OEM.ID != "rackspace" {
		t.Errorf("Failed parsing coreos.oem. Expected ID 'rackspace', got %q.", cfg.Coreos.OEM.ID)
	}

	if cfg.Hostname != "trontastic" {
		t.Errorf("Failed to parse hostname")
	}
	if cfg.Coreos.Update.RebootStrategy != "reboot" {
		t.Errorf("Failed to parse locksmith strategy")
	}

	contents = `
coreos:
write_files:
  - path: /home/me/notes
    permissions: 0744
`
	cfg, err = NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error :%v", err)
	}

	if len(cfg.WriteFiles) != 1 {
		t.Error("Failed to parse correct number of write_files")
	} else {
		wf := cfg.WriteFiles[0]
		if wf.Content != "" {
			t.Errorf("WriteFile has incorrect contents '%s'", wf.Content)
		}
		if wf.Encoding != "" {
			t.Errorf("WriteFile has incorrect encoding %s", wf.Encoding)
		}
		// Verify that the normalization of the config converted 0744 to its decimal
		// representation, 484.
		if wf.RawFilePermissions != "484" {
			t.Errorf("WriteFile has incorrect permissions %s", wf.RawFilePermissions)
		}
		if wf.Path != "/home/me/notes" {
			t.Errorf("WriteFile has incorrect path %s", wf.Path)
		}
		if wf.Owner != "" {
			t.Errorf("WriteFile has incorrect owner %s", wf.Owner)
		}
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

func TestCloudConfigUsers(t *testing.T) {
	contents := `
users:
  - name: elroy
    passwd: somehash
    ssh_authorized_keys:
      - somekey
    gecos: arbitrary comment
    homedir: /home/place
    no_create_home: yes
    primary_group: things
    groups:
      - ping
      - pong
    no_user_group: true
    system: y
    no_log_init: True
`
	cfg, err := NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error: %v", err)
	}

	if len(cfg.Users) != 1 {
		t.Fatalf("Parsed %d users, expected 1", len(cfg.Users))
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
		t.Errorf("Failed to parse no_create_home field")
	}

	if user.PrimaryGroup != "things" {
		t.Errorf("Failed to parse primary_group field, got %q", user.PrimaryGroup)
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
		t.Errorf("Failed to parse no_user_group field")
	}

	if !user.System {
		t.Errorf("Failed to parse system field")
	}

	if !user.NoLogInit {
		t.Errorf("Failed to parse no_log_init field")
	}
}

func TestCloudConfigUsersGithubUser(t *testing.T) {

	contents := `
users:
  - name: elroy
    coreos_ssh_import_github: bcwaldon
`
	cfg, err := NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error: %v", err)
	}

	if len(cfg.Users) != 1 {
		t.Fatalf("Parsed %d users, expected 1", len(cfg.Users))
	}

	user := cfg.Users[0]

	if user.Name != "elroy" {
		t.Errorf("User name is %q, expected 'elroy'", user.Name)
	}

	if user.SSHImportGithubUser != "bcwaldon" {
		t.Errorf("github user is %q, expected 'bcwaldon'", user.SSHImportGithubUser)
	}
}

func TestCloudConfigUsersSSHImportURL(t *testing.T) {
	contents := `
users:
  - name: elroy
    coreos_ssh_import_url: https://token:x-auth-token@github.enterprise.com/api/v3/polvi/keys
`
	cfg, err := NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error: %v", err)
	}

	if len(cfg.Users) != 1 {
		t.Fatalf("Parsed %d users, expected 1", len(cfg.Users))
	}

	user := cfg.Users[0]

	if user.Name != "elroy" {
		t.Errorf("User name is %q, expected 'elroy'", user.Name)
	}

	if user.SSHImportURL != "https://token:x-auth-token@github.enterprise.com/api/v3/polvi/keys" {
		t.Errorf("ssh import url is %q, expected 'https://token:x-auth-token@github.enterprise.com/api/v3/polvi/keys'", user.SSHImportURL)
	}
}

func TestNormalizeKeys(t *testing.T) {
	for _, tt := range []struct {
		in  string
		out string
	}{
		{"my_key_name: the-value\n", "my_key_name: the-value\n"},
		{"my-key_name: the-value\n", "my_key_name: the-value\n"},
		{"my-key-name: the-value\n", "my_key_name: the-value\n"},

		{"a:\n- key_name: the-value\n", "a:\n- key_name: the-value\n"},
		{"a:\n- key-name: the-value\n", "a:\n- key_name: the-value\n"},

		{"a:\n  b:\n  - key_name: the-value\n", "a:\n  b:\n  - key_name: the-value\n"},
		{"a:\n  b:\n  - key-name: the-value\n", "a:\n  b:\n  - key_name: the-value\n"},

		{"coreos:\n  update:\n    reboot-strategy: off\n", "coreos:\n  update:\n    reboot_strategy: false\n"},
		{"coreos:\n  update:\n    reboot-strategy: 'off'\n", "coreos:\n  update:\n    reboot_strategy: \"off\"\n"},
	} {
		out, err := normalizeConfig(tt.in)
		if err != nil {
			t.Fatalf("bad error (%q): want nil, got %s", tt.in, err)
		}
		if string(out) != tt.out {
			t.Fatalf("bad normalization (%q): want %q, got %q", tt.in, tt.out, out)
		}
	}
}
