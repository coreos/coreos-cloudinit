package initialize

import (
	"encoding/json"
	"fmt"

	"github.com/coreos/coreos-cloudinit/httpbackoff"
	"github.com/coreos/coreos-cloudinit/system"
)

type UserKey struct {
	ID  int    `json:"id,omitempty"`
	Key string `json:"key"`
}

func SSHImportKeysFromURL(system_user string, url string) error {
	keys, err := fetchUserKeys(url)
	if err != nil {
		return err
	}

	key_name := fmt.Sprintf("coreos-cloudinit-%s", system_user)
	return system.AuthorizeSSHKeys(system_user, key_name, keys)
}

func fetchUserKeys(url string) ([]string, error) {
	client := httpbackoff.NewHttpClient()
	data, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	var userKeys []UserKey
	err = json.Unmarshal(data, &userKeys)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0)
	for _, key := range userKeys {
		keys = append(keys, key.Key)
	}
	return keys, err
}
