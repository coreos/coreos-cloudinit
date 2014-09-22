package system

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func testReadConfig(config string) func() (io.Reader, error) {
	return func() (io.Reader, error) {
		return strings.NewReader(config), nil
	}
}

func TestUpdateUnits(t *testing.T) {
	for _, tt := range []struct {
		config config.Update
		units  []Unit
		err    error
	}{
		{
			config: config.Update{},
		},
		{
			config: config.Update{Group: "master", Server: "http://foo.com"},
			units: []Unit{{config.Unit{
				Name:    "update-engine.service",
				Command: "restart",
			}}},
		},
		{
			config: config.Update{RebootStrategy: "best-effort"},
			units: []Unit{{config.Unit{
				Name:    "locksmithd.service",
				Command: "restart",
				Runtime: true,
			}}},
		},
		{
			config: config.Update{RebootStrategy: "etcd-lock"},
			units: []Unit{{config.Unit{
				Name:    "locksmithd.service",
				Command: "restart",
				Runtime: true,
			}}},
		},
		{
			config: config.Update{RebootStrategy: "reboot"},
			units: []Unit{{config.Unit{
				Name:    "locksmithd.service",
				Command: "restart",
				Runtime: true,
			}}},
		},
		{
			config: config.Update{RebootStrategy: "off"},
			units: []Unit{{config.Unit{
				Name:    "locksmithd.service",
				Command: "stop",
				Runtime: true,
				Mask:    true,
			}}},
		},
	} {
		units, err := Update{tt.config, testReadConfig("")}.Units()
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (%q): want %q, got %q", tt.config, tt.err, err)
		}
		if !reflect.DeepEqual(tt.units, units) {
			t.Errorf("bad units (%q): want %#v, got %#v", tt.config, tt.units, units)
		}
	}
}

func TestUpdateFile(t *testing.T) {
	for _, tt := range []struct {
		config config.Update
		orig   string
		file   *File
		err    error
	}{
		{
			config: config.Update{},
		},
		{
			config: config.Update{RebootStrategy: "wizzlewazzle"},
			err:    errors.New("invalid value \"wizzlewazzle\" for option \"RebootStrategy\" (valid options: \"best-effort,etcd-lock,reboot,off\")"),
		},
		{
			config: config.Update{Group: "master", Server: "http://foo.com"},
			file: &File{
				Content:            "GROUP=master\nSERVER=http://foo.com\n",
				Path:               "etc/coreos/update.conf",
				RawFilePermissions: "0644",
			},
		},
		{
			config: config.Update{RebootStrategy: "best-effort"},
			file: &File{
				Content:            "REBOOT_STRATEGY=best-effort\n",
				Path:               "etc/coreos/update.conf",
				RawFilePermissions: "0644",
			},
		},
		{
			config: config.Update{RebootStrategy: "etcd-lock"},
			file: &File{
				Content:            "REBOOT_STRATEGY=etcd-lock\n",
				Path:               "etc/coreos/update.conf",
				RawFilePermissions: "0644",
			},
		},
		{
			config: config.Update{RebootStrategy: "reboot"},
			file: &File{
				Content:            "REBOOT_STRATEGY=reboot\n",
				Path:               "etc/coreos/update.conf",
				RawFilePermissions: "0644",
			},
		},
		{
			config: config.Update{RebootStrategy: "off"},
			file: &File{
				Content:            "REBOOT_STRATEGY=off\n",
				Path:               "etc/coreos/update.conf",
				RawFilePermissions: "0644",
			},
		},
		{
			config: config.Update{RebootStrategy: "etcd-lock"},
			orig:   "SERVER=https://example.com\nGROUP=thegroupc\nREBOOT_STRATEGY=awesome",
			file: &File{
				Content:            "SERVER=https://example.com\nGROUP=thegroupc\nREBOOT_STRATEGY=etcd-lock\n",
				Path:               "etc/coreos/update.conf",
				RawFilePermissions: "0644",
			},
		},
	} {
		file, err := Update{tt.config, testReadConfig(tt.orig)}.File()
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (%q): want %q, got %q", tt.config, tt.err, err)
		}
		if !reflect.DeepEqual(tt.file, file) {
			t.Errorf("bad units (%q): want %#v, got %#v", tt.config, tt.file, file)
		}
	}
}
