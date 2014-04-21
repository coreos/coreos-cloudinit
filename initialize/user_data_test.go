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
