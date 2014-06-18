package datasource

import (
	"io/ioutil"
	"os"
	"path"
)

type configDrive struct {
	root string
}

func NewConfigDrive(root string) *configDrive {
	return &configDrive{path.Join(root, "openstack")}
}

func (self *configDrive) ConfigRoot() string {
	return self.root
}

func (self *configDrive) Fetch() ([]byte, error) {
	return self.readFile("user_data")
}

func (self *configDrive) Type() string {
	return "cloud-drive"
}

func (self *configDrive) readFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(path.Join(self.root, "latest", filename))
	if os.IsNotExist(err) {
		err = nil
	}
	return data, err
}
