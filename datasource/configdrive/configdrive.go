package configdrive

import (
	"io/ioutil"
	"os"
	"path"
)

const (
	ec2ApiVersion       = "2009-04-04"
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

// FetchMetadata attempts to retrieve metadata from ec2/2009-04-04/meta_data.json.
func (cd *configDrive) FetchMetadata() ([]byte, error) {
	return cd.tryReadFile(path.Join(cd.ec2Root(), "meta_data.json"))
}

// FetchUserdata attempts to retrieve the userdata from ec2/2009-04-04/user_data.
// If no data is found, it will attempt to read from openstack/latest/user_data.
func (cd *configDrive) FetchUserdata() ([]byte, error) {
	bytes, err := cd.tryReadFile(path.Join(cd.ec2Root(), "user_data"))
	if bytes == nil && err == nil {
		bytes, err = cd.tryReadFile(path.Join(cd.openstackRoot(), "user_data"))
	}
	return bytes, err
}

func (cd *configDrive) Type() string {
	return "cloud-drive"
}

func (cd *configDrive) ec2Root() string {
	return path.Join(cd.root, "ec2", ec2ApiVersion)
}

func (cd *configDrive) openstackRoot() string {
	return path.Join(cd.root, "openstack", openstackApiVersion)
}

func (cd *configDrive) tryReadFile(filename string) ([]byte, error) {
	data, err := cd.readFile(filename)
	if os.IsNotExist(err) {
		err = nil
	}
	return data, err
}
