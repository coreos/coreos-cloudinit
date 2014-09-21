package main

import (
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func TestMergeCloudConfig(t *testing.T) {
	simplecc := config.CloudConfig{
		SSHAuthorizedKeys: []string{"abc", "def"},
		Hostname:          "foobar",
		NetworkConfigPath: "/path/somewhere",
		NetworkConfig:     `{}`,
	}
	for i, tt := range []struct {
		udcc config.CloudConfig
		mdcc config.CloudConfig
		want config.CloudConfig
	}{
		{
			// If mdcc is empty, udcc should be returned unchanged
			simplecc,
			config.CloudConfig{},
			simplecc,
		},
		{
			// If udcc is empty, mdcc should be returned unchanged(overridden)
			config.CloudConfig{},
			simplecc,
			simplecc,
		},
		{
			// user-data should override completely in the case of conflicts
			simplecc,
			config.CloudConfig{
				Hostname:          "meta-hostname",
				NetworkConfigPath: "/path/meta",
				NetworkConfig:     `{"hostname":"test"}`,
			},
			simplecc,
		},
		{
			// Mixed merge should succeed
			config.CloudConfig{
				SSHAuthorizedKeys: []string{"abc", "def"},
				Hostname:          "user-hostname",
				NetworkConfigPath: "/path/somewhere",
				NetworkConfig:     `{"hostname":"test"}`,
			},
			config.CloudConfig{
				SSHAuthorizedKeys: []string{"woof", "qux"},
				Hostname:          "meta-hostname",
			},
			config.CloudConfig{
				SSHAuthorizedKeys: []string{"abc", "def", "woof", "qux"},
				Hostname:          "user-hostname",
				NetworkConfigPath: "/path/somewhere",
				NetworkConfig:     `{"hostname":"test"}`,
			},
		},
		{
			// Completely non-conflicting merge should be fine
			config.CloudConfig{
				Hostname: "supercool",
			},
			config.CloudConfig{
				SSHAuthorizedKeys: []string{"zaphod", "beeblebrox"},
				NetworkConfigPath: "/dev/fun",
				NetworkConfig:     `{"hostname":"test"}`,
			},
			config.CloudConfig{
				Hostname:          "supercool",
				SSHAuthorizedKeys: []string{"zaphod", "beeblebrox"},
				NetworkConfigPath: "/dev/fun",
				NetworkConfig:     `{"hostname":"test"}`,
			},
		},
		{
			// Non-mergeable settings in user-data should not be affected
			config.CloudConfig{
				Hostname:       "mememe",
				ManageEtcHosts: config.EtcHosts("lolz"),
			},
			config.CloudConfig{
				Hostname:          "youyouyou",
				NetworkConfigPath: "meta-meta-yo",
				NetworkConfig:     `{"hostname":"test"}`,
			},
			config.CloudConfig{
				Hostname:          "mememe",
				ManageEtcHosts:    config.EtcHosts("lolz"),
				NetworkConfigPath: "meta-meta-yo",
				NetworkConfig:     `{"hostname":"test"}`,
			},
		},
		{
			// Non-mergeable (unexpected) settings in meta-data are ignored
			config.CloudConfig{
				Hostname: "mememe",
			},
			config.CloudConfig{
				ManageEtcHosts:    config.EtcHosts("lolz"),
				NetworkConfigPath: "meta-meta-yo",
				NetworkConfig:     `{"hostname":"test"}`,
			},
			config.CloudConfig{
				Hostname:          "mememe",
				NetworkConfigPath: "meta-meta-yo",
				NetworkConfig:     `{"hostname":"test"}`,
			},
		},
	} {
		got := mergeCloudConfig(tt.mdcc, tt.udcc)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("case #%d: mergeCloudConfig mutated CloudConfig unexpectedly:\ngot:\n%s\nwant:\n%s", i, got, tt.want)
		}
	}
}
