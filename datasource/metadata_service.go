package datasource

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
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
	Ec2MetadataUrl       = BaseUrl + Ec2ApiVersion + "/meta-data"
	OpenstackApiVersion  = "openstack/2012-08-10"
	OpenstackUserdataUrl = BaseUrl + OpenstackApiVersion + "/user_data"
	OpenstackMetadataUrl = BaseUrl + OpenstackApiVersion + "/meta_data.json"
)

type metadataService struct{}

type getter interface {
	GetRetry(string) ([]byte, error)
}

func NewMetadataService() *metadataService {
	return &metadataService{}
}

func (ms *metadataService) IsAvailable() bool {
	client := pkg.NewHttpClient()
	_, err := client.Get(BaseUrl)
	return (err == nil)
}

func (ms *metadataService) AvailabilityChanges() bool {
	return true
}

func (ms *metadataService) ConfigRoot() string {
	return ""
}

func (ms *metadataService) FetchMetadata() ([]byte, error) {
	return fetchMetadata(pkg.NewHttpClient())
}

func (ms *metadataService) FetchUserdata() ([]byte, error) {
	client := pkg.NewHttpClient()
	if data, err := client.GetRetry(Ec2UserdataUrl); err == nil {
		return data, err
	} else if _, ok := err.(pkg.ErrTimeout); ok {
		return data, err
	}
	return client.GetRetry(OpenstackUserdataUrl)
}

func (ms *metadataService) Type() string {
	return "metadata-service"
}

func fetchMetadata(client getter) ([]byte, error) {
	if metadata, err := client.GetRetry(OpenstackMetadataUrl); err == nil {
		return metadata, nil
	} else if _, ok := err.(pkg.ErrTimeout); ok {
		return nil, err
	}

	attrs := make(map[string]interface{})
	if keynames, err := fetchAttributes(client, fmt.Sprintf("%s/public-keys", Ec2MetadataUrl)); err == nil {
		keyIDs := make(map[string]string)
		for _, keyname := range keynames {
			tokens := strings.SplitN(keyname, "=", 2)
			if len(tokens) != 2 {
				return nil, fmt.Errorf("malformed public key: %q\n", keyname)
			}
			keyIDs[tokens[1]] = tokens[0]
		}

		keys := make(map[string]string)
		for name, id := range keyIDs {
			sshkey, err := fetchAttribute(client, fmt.Sprintf("%s/public-keys/%s/openssh-key", Ec2MetadataUrl, id))
			if err != nil {
				return nil, err
			}
			keys[name] = sshkey
			fmt.Printf("Found SSH key for %q\n", name)
		}
		attrs["public_keys"] = keys
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if hostname, err := fetchAttribute(client, fmt.Sprintf("%s/hostname", Ec2MetadataUrl)); err == nil {
		attrs["hostname"] = hostname
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if content_path, err := fetchAttribute(client, fmt.Sprintf("%s/network_config/content_path", Ec2MetadataUrl)); err == nil {
		attrs["network_config"] = map[string]string{
			"content_path": content_path,
		}
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	return json.Marshal(attrs)
}

func fetchAttributes(client getter, url string) ([]string, error) {
	resp, err := client.GetRetry(url)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(resp))
	data := make([]string, 0)
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}
	return data, scanner.Err()
}

func fetchAttribute(client getter, url string) (string, error) {
	if attrs, err := fetchAttributes(client, url); err == nil && len(attrs) > 0 {
		return attrs[0], nil
	} else {
		return "", err
	}
}
