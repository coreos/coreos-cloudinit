package initialize

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var data []UserKey
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0)
	for _, key := range data {
		keys = append(keys, key.Key)
	}
	return keys, err
}
