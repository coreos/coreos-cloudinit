package configdrive

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

const (
	openstackApiVersion = "latest"
)

type configDrive struct {
	root     string
	readFile func(filename string) ([]byte, error)
}

func NewDatasource(root string) *configDrive {
	return &configDrive{root, ioutil.ReadFile}
}

func (cd *configDrive) IsAvailable() bool {
	_, err := os.Stat(cd.root)
	return !os.IsNotExist(err)
}

func (cd *configDrive) AvailabilityChanges() bool {
	return true
}

func (cd *configDrive) ConfigRoot() string {
	return cd.openstackRoot()
}

func (cd *configDrive) FetchMetadata() ([]byte, error) {
	return cd.tryReadFile(path.Join(cd.openstackVersionRoot(), "meta_data.json"))
}

func (cd *configDrive) FetchUserdata() ([]byte, error) {
	return cd.tryReadFile(path.Join(cd.openstackVersionRoot(), "user_data"))
}

func (cd *configDrive) FetchNetworkConfig(filename string) ([]byte, error) {
	if filename == "" {
		return []byte{}, nil
	}
	return cd.tryReadFile(path.Join(cd.openstackRoot(), filename))
}

func (cd *configDrive) Type() string {
	return "cloud-drive"
}

func (cd *configDrive) openstackRoot() string {
	return path.Join(cd.root, "openstack")
}

func (cd *configDrive) openstackVersionRoot() string {
	return path.Join(cd.openstackRoot(), openstackApiVersion)
}

func (cd *configDrive) tryReadFile(filename string) ([]byte, error) {
	fmt.Printf("Attempting to read from %q\n", filename)
	data, err := cd.readFile(filename)
	if os.IsNotExist(err) {
		err = nil
	}
	return data, err
}
