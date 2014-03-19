package initialize

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/coreos/coreos-cloudinit/system"
)

type GithubUserKey struct {
	Id  int    `json:"id"`
	Key string `json:"key"`
}

func fetchGithubKeys(github_url string) ([]string, error) {
	res, err := http.Get(github_url)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var data []GithubUserKey
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

func SSHImportGithubUser(system_user string, github_user string) error {
	url := fmt.Sprintf("https://api.github.com/users/%s/keys", github_user)
	keys, err := fetchGithubKeys(url)
	if err != nil {
		return err
	}
	key_name := fmt.Sprintf("github-%s", github_user)
	err = system.AuthorizeSSHKeys(system_user, key_name, keys)
	if err != nil {
		return err
	}
	return nil
}
