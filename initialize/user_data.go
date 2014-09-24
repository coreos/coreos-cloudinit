package initialize

import (
	"fmt"
	"log"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/system"
)

func ParseUserData(contents string) (interface{}, error) {
	if len(contents) == 0 {
		return nil, nil
	}
	header := strings.SplitN(contents, "\n", 2)[0]

	// Explicitly trim the header so we can handle user-data from
	// non-unix operating systems. The rest of the file is parsed
	// by yaml, which correctly handles CRLF.
	header = strings.TrimSpace(header)

	if strings.HasPrefix(header, "#!") {
		log.Printf("Parsing user-data as script")
		return system.Script(contents), nil
	} else if header == "#cloud-config" {
		log.Printf("Parsing user-data as cloud-config")
		return config.NewCloudConfig(contents)
	} else {
		return nil, fmt.Errorf("Unrecognized user-data header: %s", header)
	}
}
