package cloudinit

import (
	"io/ioutil"
	"net/http"
)

type metadataService struct {
	client http.Client
}

func NewMetadataService() *metadataService {
	return &metadataService{http.Client{}}
}

func (ms *metadataService) UserData() ([]byte, error) {
	resp, err := ms.client.Get("http://169.254.169.254/2012-01-12/user-data")
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBytes, nil
}


