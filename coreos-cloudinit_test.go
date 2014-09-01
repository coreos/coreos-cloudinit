package main

import (
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/initialize"
)

func TestMergeCloudConfig(t *testing.T) {
	simplecc := initialize.CloudConfig{
		SSHAuthorizedKeys: []string{"abc", "def"},
		Hostname:          "foobar",
		NetworkConfigPath: "/path/somewhere",
		NetworkConfig:     `{}`,
	}
	for i, tt := range []struct {
		udcc initialize.CloudConfig
		mdcc initialize.CloudConfig
		want initialize.CloudConfig
	}{
		{
			// If mdcc is empty, udcc should be returned unchanged
			simplecc,
			initialize.CloudConfig{},
			simplecc,
		},
		{
			// If udcc is empty, mdcc should be returned unchanged(overridden)
			initialize.CloudConfig{},
			simplecc,
			simplecc,
		},
		{
			// user-data should override completely in the case of conflicts
			simplecc,
			initialize.CloudConfig{
				Hostname:          "meta-hostname",
				NetworkConfigPath: "/path/meta",
				NetworkConfig:     `{"hostname":"test"}`,
			},
			simplecc,
		},
		{
			// Mixed merge should succeed
			initialize.CloudConfig{
				SSHAuthorizedKeys: []string{"abc", "def"},
				Hostname:          "user-hostname",
				NetworkConfigPath: "/path/somewhere",
				NetworkConfig:     `{"hostname":"test"}`,
			},
			initialize.CloudConfig{
				SSHAuthorizedKeys: []string{"woof", "qux"},
				Hostname:          "meta-hostname",
			},
			initialize.CloudConfig{
				SSHAuthorizedKeys: []string{"abc", "def", "woof", "qux"},
				Hostname:          "user-hostname",
				NetworkConfigPath: "/path/somewhere",
				NetworkConfig:     `{"hostname":"test"}`,
			},
		},
		{
			// Completely non-conflicting merge should be fine
			initialize.CloudConfig{
				Hostname: "supercool",
			},
			initialize.CloudConfig{
				SSHAuthorizedKeys: []string{"zaphod", "beeblebrox"},
				NetworkConfigPath: "/dev/fun",
				NetworkConfig:     `{"hostname":"test"}`,
			},
			initialize.CloudConfig{
				Hostname:          "supercool",
				SSHAuthorizedKeys: []string{"zaphod", "beeblebrox"},
				NetworkConfigPath: "/dev/fun",
				NetworkConfig:     `{"hostname":"test"}`,
			},
		},
		{
			// Non-mergeable settings in user-data should not be affected
			initialize.CloudConfig{
				Hostname:       "mememe",
				ManageEtcHosts: initialize.EtcHosts("lolz"),
			},
			initialize.CloudConfig{
				Hostname:          "youyouyou",
				NetworkConfigPath: "meta-meta-yo",
				NetworkConfig:     `{"hostname":"test"}`,
			},
			initialize.CloudConfig{
				Hostname:          "mememe",
				ManageEtcHosts:    initialize.EtcHosts("lolz"),
				NetworkConfigPath: "meta-meta-yo",
				NetworkConfig:     `{"hostname":"test"}`,
			},
		},
		{
			// Non-mergeable (unexpected) settings in meta-data are ignored
			initialize.CloudConfig{
				Hostname: "mememe",
			},
			initialize.CloudConfig{
				ManageEtcHosts:    initialize.EtcHosts("lolz"),
				NetworkConfigPath: "meta-meta-yo",
				NetworkConfig:     `{"hostname":"test"}`,
			},
			initialize.CloudConfig{
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
