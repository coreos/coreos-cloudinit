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

func (cd *configDrive) ConfigRoot() string {
	return cd.root
}

func (cd *configDrive) FetchMetadata() ([]byte, error) {
	return cd.readFile("meta_data.json")
}

func (cd *configDrive) FetchUserdata() ([]byte, error) {
	return cd.readFile("user_data")
}

func (cd *configDrive) Type() string {
	return "cloud-drive"
}

func (cd *configDrive) readFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(path.Join(cd.root, "latest", filename))
	if os.IsNotExist(err) {
		err = nil
	}
	return data, err
}
