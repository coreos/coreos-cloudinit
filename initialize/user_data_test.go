package initialize

import (
	"testing"
)

func TestParseHeaderCRLF(t *testing.T) {
	configs := []string{
		"#cloud-config\nfoo: bar",
		"#cloud-config\r\nfoo: bar",
	}

	for i, config := range configs {
		_, err := ParseUserData(config)
		if err != nil {
			t.Errorf("Failed parsing config %d: %v", i, err)
		}
	}

	scripts := []string{
		"#!bin/bash\necho foo",
		"#!bin/bash\r\necho foo",
	}

	for i, script := range scripts {
		_, err := ParseUserData(script)
		if err != nil {
			t.Errorf("Failed parsing script %d: %v", i, err)
		}
	}
}

func TestParseConfigCRLF(t *testing.T) {
	contents := "#cloud-config\r\nhostname: foo\r\nssh_authorized_keys:\r\n  - foobar\r\n"
	ud, err := ParseUserData(contents)
	if err != nil {
		t.Fatalf("Failed parsing config: %v", err)
	}

	cfg := ud.(*CloudConfig)

	if cfg.Hostname != "foo" {
		t.Error("Failed parsing hostname from config")
	}

	if len(cfg.SSHAuthorizedKeys) != 1 {
		t.Error("Parsed incorrect number of SSH keys")
	}
}

func TestParseConfigEmpty(t *testing.T) {
	i, e := ParseUserData(``)
	if i != nil {
		t.Error("ParseUserData of empty string returned non-nil unexpectedly")
	} else if e != nil {
		t.Error("ParseUserData of empty string returned error unexpectedly")
	}
}
