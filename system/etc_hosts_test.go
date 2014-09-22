package system

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func TestEtcdHostsFile(t *testing.T) {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	for _, tt := range []struct {
		config config.EtcHosts
		file   *File
		err    error
	}{
		{
			"invalid",
			nil,
			fmt.Errorf("Invalid option to manage_etc_hosts"),
		},
		{
			"localhost",
			&File{
				Content:            fmt.Sprintf("127.0.0.1 %s\n", hostname),
				Path:               "etc/hosts",
				RawFilePermissions: "0644",
			},
			nil,
		},
	} {
		file, err := EtcHosts{tt.config}.File()
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (%q): want %q, got %q", tt.config, tt.err, err)
		}
		if !reflect.DeepEqual(tt.file, file) {
			t.Errorf("bad units (%q): want %#v, got %#v", tt.config, tt.file, file)
		}
	}
}
