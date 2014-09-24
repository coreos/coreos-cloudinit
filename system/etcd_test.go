package system

import (
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func TestEtcdUnits(t *testing.T) {
	for _, tt := range []struct {
		config config.Etcd
		units  []Unit
	}{
		{
			config.Etcd{},
			nil,
		},
		{
			config.Etcd{
				Discovery:    "http://disco.example.com/foobar",
				PeerBindAddr: "127.0.0.1:7002",
			},
			[]Unit{{config.Unit{
				Name:    "etcd.service",
				Runtime: true,
				DropIn:  true,
				Content: `[Service]
Environment="ETCD_DISCOVERY=http://disco.example.com/foobar"
Environment="ETCD_PEER_BIND_ADDR=127.0.0.1:7002"
`,
			}}},
		},
		{
			config.Etcd{
				Name:         "node001",
				Discovery:    "http://disco.example.com/foobar",
				PeerBindAddr: "127.0.0.1:7002",
			},
			[]Unit{{config.Unit{
				Name:    "etcd.service",
				Runtime: true,
				DropIn:  true,
				Content: `[Service]
Environment="ETCD_DISCOVERY=http://disco.example.com/foobar"
Environment="ETCD_NAME=node001"
Environment="ETCD_PEER_BIND_ADDR=127.0.0.1:7002"
`,
			}}},
		},
	} {
		units, err := Etcd{tt.config}.Units()
		if err != nil {
			t.Errorf("bad error (%q): want %q, got %q", tt.config, nil, err)
		}
		if !reflect.DeepEqual(tt.units, units) {
			t.Errorf("bad units (%q): want %#v, got %#v", tt.config, tt.units, units)
		}
	}
}
