package initialize

import (
	"fmt"
	"path"
	"strings"

	"github.com/coreos/coreos-cloudinit/system"
)

type OEMRelease struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	VersionID    string `yaml:"version-id"`
	HomeURL      string `yaml:"home-url"`
	BugReportURL string `yaml:"bug-report-url"`
}

func (oem *OEMRelease) String() string {
	fields := []string{
		fmt.Sprintf("ID=%s", oem.ID),
		fmt.Sprintf("VERSION_ID=%s", oem.VersionID),
		fmt.Sprintf("NAME=%q", oem.Name),
		fmt.Sprintf("HOME_URL=%q", oem.HomeURL),
		fmt.Sprintf("BUG_REPORT_URL=%q", oem.BugReportURL),
	}

	return strings.Join(fields, "\n") + "\n"
}

func WriteOEMRelease(oem *OEMRelease, root string) error {
	file := system.File{
		Path:               path.Join(root, "etc", "oem-release"),
		RawFilePermissions: "0644",
		Content:            oem.String(),
	}

	return system.WriteFile(&file)
}
