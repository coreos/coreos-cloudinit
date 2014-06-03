package datasource

import (
	"io/ioutil"
	"os"
	"path"
)

type configDrive struct {
	path string
}

func NewConfigDrive(path string) *configDrive {
	return &configDrive{path}
}

func (self *configDrive) Fetch() ([]byte, error) {
	data, err := ioutil.ReadFile(path.Join(self.path, "openstack", "latest", "user_data"))
	if os.IsNotExist(err) {
		err = nil
	}
	return data, err
}

func (self *configDrive) Type() string {
	return "cloud-drive"
}
