package datasource

import (
	"errors"
	"io/ioutil"
	"log"
	"strings"

	"github.com/coreos/coreos-cloudinit/httpbackoff"
)

const (
	ProcCmdlineLocation        = "/proc/cmdline"
	ProcCmdlineCloudConfigFlag = "cloud-config-url"
)

type procCmdline struct{}

func NewProcCmdline() *procCmdline {
	return &procCmdline{}
}

func (self *procCmdline) Fetch() ([]byte, error) {
	cmdline, err := ioutil.ReadFile(ProcCmdlineLocation)
	if err != nil {
		return nil, err
	}

	url, err := findCloudConfigURL(string(cmdline))
	if err != nil {
		return nil, err
	}

	client := httpbackoff.NewHttpClient()
	cfg, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (self *procCmdline) Type() string {
	return "proc-cmdline"
}

func findCloudConfigURL(input string) (url string, err error) {
	err = errors.New("cloud-config-url not found")
	for _, token := range strings.Split(input, " ") {
		parts := strings.SplitN(token, "=", 2)

		key := parts[0]
		key = strings.Replace(key, "_", "-", -1)

		if key != "cloud-config-url" {
			continue
		}

		if len(parts) != 2 {
			log.Printf("Found cloud-config-url in /proc/cmdline with no value, ignoring.")
			continue
		}

		url = parts[1]
		err = nil
	}

	return
}
