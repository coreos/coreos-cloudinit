package datasource

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"

	"github.com/coreos/coreos-cloudinit/pkg"
)

// metadataService retrieves metadata from either an OpenStack[1] or EC2[2] compatible endpoint.
// It will first attempt to directly retrieve a JSON blob from the OpenStack endpoint. If that
// fails with a 404, it then attempts to retrieve metadata bit-by-bit from the EC2 endpoint,
// and populates that into an equivalent JSON blob.
//
// [1] http://docs.openstack.org/grizzly/openstack-compute/admin/content/metadata-service.html
// [2] http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AESDG-chapter-instancedata.html#instancedata-data-categories
type metadataService struct {
	url string
}

type getter interface {
	Get(string) ([]byte, error)
}

func NewMetadataService(url string) *metadataService {
	return &metadataService{strings.TrimSuffix(url, "/")}
}

func (ms *metadataService) ConfigRoot() string {
	return ""
}

func (ms *metadataService) FetchMetadata() ([]byte, error) {
	client := pkg.NewHttpClient()
	return fetchMetadata(client, ms.url)
}

func (ms *metadataService) FetchUserdata() ([]byte, error) {
	client := pkg.NewHttpClient()
	return client.Get(ms.url + "/latest/user-data")
}

func (ms *metadataService) Type() string {
	return "metadata-service"
}

func fetchMetadata(client getter, url string) ([]byte, error) {
	if metadata, err := client.Get(url + "/latest/meta-data.json"); err == nil {
		return metadata, nil
	} else if _, ok := err.(pkg.ErrTimeout); ok {
		return nil, err
	}

	attrs, err := fetchChildAttributes(client, url+"/latest/meta-data/")
	if err != nil {
		return nil, err
	}
	return json.Marshal(attrs)
}

func fetchAttributes(client getter, url string) ([]string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(resp))
	data := make([]string, 0)
	for scanner.Scan() {
		data = append(data, strings.Split(scanner.Text(), "=")[0])
	}
	return data, scanner.Err()
}

func fetchAttribute(client getter, url string) (interface{}, error) {
	if attrs, err := fetchAttributes(client, url); err == nil {
		return attrs[0], nil
	} else {
		return "", err
	}
}

func fetchChildAttributes(client getter, url string) (interface{}, error) {
	attrs := make(map[string]interface{})
	attrList, err := fetchAttributes(client, url)
	if err != nil {
		return nil, err
	}
	for _, attr := range attrList {
		var fetchFunc func(getter, string) (interface{}, error)
		if strings.HasSuffix(attr, "/") {
			fetchFunc = fetchChildAttributes
		} else {
			fetchFunc = fetchAttribute
		}
		if value, err := fetchFunc(client, url+attr); err == nil {
			attrs[strings.TrimSuffix(attr, "/")] = value
		} else {
			return nil, err
		}
	}
	return attrs, nil
}
