package datasource

import "github.com/coreos/coreos-cloudinit/pkg"

type metadataService struct {
	url string
}

func NewMetadataService(url string) *metadataService {
	return &metadataService{url}
}

func (self *metadataService) ConfigRoot() string {
	return ""
}

func (self *metadataService) FetchMetadata() ([]byte, error) {
	return []byte{}, nil
}

func (ms *metadataService) FetchUserdata() ([]byte, error) {
	client := pkg.NewHttpClient()
	return client.Get(ms.url)
}

func (ms *metadataService) Type() string {
	return "metadata-service"
}
