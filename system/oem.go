package system

import (
	"fmt"
	"path"

	"github.com/coreos/coreos-cloudinit/config"
)

// OEM is a top-level structure which embeds its underlying configuration,
// config.OEM, and provides the system-specific File().
type OEM struct {
	config.OEM
}

func (oem OEM) File(_ string) (*File, error) {
	if oem.ID == "" {
		return nil, nil
	}

	content := fmt.Sprintf("ID=%s\n", oem.ID)
	content += fmt.Sprintf("VERSION_ID=%s\n", oem.VersionID)
	content += fmt.Sprintf("NAME=%q\n", oem.Name)
	content += fmt.Sprintf("HOME_URL=%q\n", oem.HomeURL)
	content += fmt.Sprintf("BUG_REPORT_URL=%q\n", oem.BugReportURL)

	return &File{
		Path:               path.Join("etc", "oem-release"),
		RawFilePermissions: "0644",
		Content:            content,
	}, nil
}
