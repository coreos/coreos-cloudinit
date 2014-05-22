package datasource

import "github.com/coreos/coreos-cloudinit/pkg"

type metadataService struct {
	url string
}

func NewMetadataService(url string) *metadataService {
	return &metadataService{url}
}

func (ms *metadataService) Fetch() ([]byte, error) {
	client := pkg.NewHttpClient()
	return client.Get(ms.url)
}

func (ms *metadataService) Type() string {
	return "metadata-service"
}
