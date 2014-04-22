package datasource

import (
	"io/ioutil"
	"net/http"
)

type Datasource interface {
	Fetch() ([]byte, error)
	Type()  string
}

func fetchURL(url string) ([]byte, error) {
	client := http.Client{}
	resp, err := client.Get(url)
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
