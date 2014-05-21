package datasource

type metadataService struct {
	url string
}

func NewMetadataService(url string) *metadataService {
	return &metadataService{url}
}

func (ms *metadataService) Fetch() ([]byte, error) {
	return fetchURL(ms.url)
}

func (ms *metadataService) Type() string {
	return "metadata-service"
}
