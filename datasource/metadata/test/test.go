package test

import (
	"fmt"

	"github.com/coreos/coreos-cloudinit/pkg"
)

type HttpClient struct {
	Resources map[string]string
	Err       error
}

func (t *HttpClient) GetRetry(url string) ([]byte, error) {
	if t.Err != nil {
		return nil, t.Err
	}
	if val, ok := t.Resources[url]; ok {
		return []byte(val), nil
	} else {
		return nil, pkg.ErrNotFound{fmt.Errorf("not found: %q", url)}
	}
}

func (t *HttpClient) Get(url string) ([]byte, error) {
	return t.GetRetry(url)
}
