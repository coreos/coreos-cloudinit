package cloudinit

import (
	"io/ioutil"
	"net/http"
)

type metadataService struct {
	url    string
	client http.Client
}

func NewMetadataService(url string) *metadataService {
	return &metadataService{url, http.Client{}}
}

func (ms *metadataService) UserData() ([]byte, error) {
	resp, err := ms.client.Get(ms.url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode / 100 != 2 {
		return []byte{}, nil
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBytes, nil
}


