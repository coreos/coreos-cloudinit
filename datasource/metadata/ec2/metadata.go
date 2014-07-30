package ec2

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coreos/coreos-cloudinit/pkg"
)

const (
	DefaultAddress = "http://169.254.169.254/"
	apiVersion     = "2009-04-04"
	userdataUrl    = apiVersion + "/user-data"
	metadataUrl    = apiVersion + "/meta-data"
)

type metadataService struct {
	root   string
	client pkg.Getter
}

func NewDatasource(root string) *metadataService {
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}
	return &metadataService{root, pkg.NewHttpClient()}
}

func (ms metadataService) IsAvailable() bool {
	_, err := ms.client.Get(ms.root + apiVersion)
	return (err == nil)
}

func (ms metadataService) AvailabilityChanges() bool {
	return true
}

func (ms metadataService) ConfigRoot() string {
	return ms.root
}

func (ms metadataService) FetchMetadata() ([]byte, error) {
	attrs := make(map[string]interface{})
	if keynames, err := fetchAttributes(ms.client, fmt.Sprintf("%s/public-keys", ms.metadataUrl())); err == nil {
		keyIDs := make(map[string]string)
		for _, keyname := range keynames {
			tokens := strings.SplitN(keyname, "=", 2)
			if len(tokens) != 2 {
				return nil, fmt.Errorf("malformed public key: %q", keyname)
			}
			keyIDs[tokens[1]] = tokens[0]
		}

		keys := make(map[string]string)
		for name, id := range keyIDs {
			sshkey, err := fetchAttribute(ms.client, fmt.Sprintf("%s/public-keys/%s/openssh-key", ms.metadataUrl(), id))
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

	if hostname, err := fetchAttribute(ms.client, fmt.Sprintf("%s/hostname", ms.metadataUrl())); err == nil {
		attrs["hostname"] = hostname
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if localAddr, err := fetchAttribute(ms.client, fmt.Sprintf("%s/local-ipv4", ms.metadataUrl())); err == nil {
		attrs["local-ipv4"] = localAddr
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if publicAddr, err := fetchAttribute(ms.client, fmt.Sprintf("%s/public-ipv4", ms.metadataUrl())); err == nil {
		attrs["public-ipv4"] = publicAddr
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if content_path, err := fetchAttribute(ms.client, fmt.Sprintf("%s/network_config/content_path", ms.metadataUrl())); err == nil {
		attrs["network_config"] = map[string]string{
			"content_path": content_path,
		}
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	return json.Marshal(attrs)
}

func (ms metadataService) FetchUserdata() ([]byte, error) {
	if data, err := ms.client.GetRetry(ms.userdataUrl()); err == nil {
		return data, err
	} else if _, ok := err.(pkg.ErrNotFound); ok {
		return []byte{}, nil
	} else {
		return data, err
	}
}

func (ms metadataService) Type() string {
	return "ec2-metadata-service"
}

func (ms metadataService) metadataUrl() string {
	return (ms.root + metadataUrl)
}

func (ms metadataService) userdataUrl() string {
	return (ms.root + userdataUrl)
}

func fetchAttributes(client pkg.Getter, url string) ([]string, error) {
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

func fetchAttribute(client pkg.Getter, url string) (string, error) {
	if attrs, err := fetchAttributes(client, url); err == nil && len(attrs) > 0 {
		return attrs[0], nil
	} else {
		return "", err
	}
}
