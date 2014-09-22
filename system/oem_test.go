package system

import (
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func TestOEMFile(t *testing.T) {
	for _, tt := range []struct {
		config config.OEM
		file   *File
	}{
		{
			config.OEM{},
			nil,
		},
		{
			config.OEM{
				ID:           "rackspace",
				Name:         "Rackspace Cloud Servers",
				VersionID:    "168.0.0",
				HomeURL:      "https://www.rackspace.com/cloud/servers/",
				BugReportURL: "https://github.com/coreos/coreos-overlay",
			},
			&File{
				Path:               "etc/oem-release",
				RawFilePermissions: "0644",
				Content: `ID=rackspace
VERSION_ID=168.0.0
NAME="Rackspace Cloud Servers"
HOME_URL="https://www.rackspace.com/cloud/servers/"
BUG_REPORT_URL="https://github.com/coreos/coreos-overlay"
`,
			},
		},
	} {
		file, err := OEM{tt.config}.File()
		if err != nil {
			t.Errorf("bad error (%q): want %q, got %q", tt.config, nil, err)
		}
		if !reflect.DeepEqual(tt.file, file) {
			t.Errorf("bad file (%q): want %#v, got %#v", tt.config, tt.file, file)
		}
	}
}
