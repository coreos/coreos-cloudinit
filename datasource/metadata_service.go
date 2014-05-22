package datasource

import "github.com/coreos/coreos-cloudinit/httpbackoff"

type metadataService struct {
	url string
}

func NewMetadataService(url string) *metadataService {
	return &metadataService{url}
}

func (ms *metadataService) Fetch() ([]byte, error) {
	client := httpbackoff.NewHttpClient()
	return client.Get(ms.url)
}

func (ms *metadataService) Type() string {
	return "metadata-service"
}
