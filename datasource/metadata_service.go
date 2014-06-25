package datasource

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"

	"github.com/coreos/coreos-cloudinit/pkg"
)

// metadataService retrieves metadata from either an OpenStack[1] (2012-08-10)
// or EC2[2] (2009-04-04) compatible endpoint. It will first attempt to
// directly retrieve a JSON blob from the OpenStack endpoint. If that fails
// with a 404, it then attempts to retrieve metadata bit-by-bit from the EC2
// endpoint, and populates that into an equivalent JSON blob. metadataService
// also checks for userdata from EC2 and, if that fails with a 404, OpenStack.
//
// [1] http://docs.openstack.org/grizzly/openstack-compute/admin/content/metadata-service.html
// [2] http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AESDG-chapter-instancedata.html#instancedata-data-categories

const (
	BaseUrl              = "http://169.254.169.254/"
	Ec2ApiVersion        = "2009-04-04"
	Ec2UserdataUrl       = BaseUrl + Ec2ApiVersion + "/user-data"
	Ec2MetadataUrl       = BaseUrl + Ec2ApiVersion + "/meta-data/"
	OpenstackApiVersion  = "openstack/2012-08-10"
	OpenstackUserdataUrl = BaseUrl + OpenstackApiVersion + "/user_data"
	OpenstackMetadataUrl = BaseUrl + OpenstackApiVersion + "/meta_data.json"
)

type metadataService struct{}

type getter interface {
	Get(string) ([]byte, error)
}

func NewMetadataService() *metadataService {
	return &metadataService{}
}

func (ms *metadataService) ConfigRoot() string {
	return ""
}

func (ms *metadataService) FetchMetadata() ([]byte, error) {
	return fetchMetadata(pkg.NewHttpClient())
}

func (ms *metadataService) FetchUserdata() ([]byte, error) {
	client := pkg.NewHttpClient()
	if data, err := client.Get(Ec2UserdataUrl); err == nil {
		return data, err
	} else if _, ok := err.(pkg.ErrTimeout); ok {
		return data, err
	}
	return client.Get(OpenstackUserdataUrl)
}

func (ms *metadataService) Type() string {
	return "metadata-service"
}

func fetchMetadata(client getter) ([]byte, error) {
	if metadata, err := client.Get(OpenstackMetadataUrl); err == nil {
		return metadata, nil
	} else if _, ok := err.(pkg.ErrTimeout); ok {
		return nil, err
	}

	attrs, err := fetchChildAttributes(client, Ec2MetadataUrl)
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
	if attrs, err := fetchAttributes(client, url); err == nil && len(attrs) > 0 {
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
