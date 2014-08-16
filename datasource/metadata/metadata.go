package metadata

import (
	"strings"

	"github.com/coreos/coreos-cloudinit/pkg"
)

type MetadataService struct {
	Root         string
	Client       pkg.Getter
	ApiVersion   string
	UserdataPath string
	MetadataPath string
}

func NewDatasource(root, apiVersion, userdataPath, metadataPath string) MetadataService {
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}
	return MetadataService{root, pkg.NewHttpClient(), apiVersion, userdataPath, metadataPath}
}

func (ms MetadataService) IsAvailable() bool {
	_, err := ms.Client.Get(ms.Root + ms.ApiVersion)
	return (err == nil)
}

func (ms MetadataService) AvailabilityChanges() bool {
	return true
}

func (ms MetadataService) ConfigRoot() string {
	return ms.Root
}

func (ms MetadataService) FetchUserdata() ([]byte, error) {
	return ms.FetchData(ms.UserdataUrl())
}

func (ms MetadataService) FetchData(url string) ([]byte, error) {
	if data, err := ms.Client.GetRetry(url); err == nil {
		return data, err
	} else if _, ok := err.(pkg.ErrNotFound); ok {
		return []byte{}, nil
	} else {
		return data, err
	}
}

func (ms MetadataService) MetadataUrl() string {
	return (ms.Root + ms.MetadataPath)
}

func (ms MetadataService) UserdataUrl() string {
	return (ms.Root + ms.UserdataPath)
}
