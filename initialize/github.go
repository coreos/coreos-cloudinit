package initialize

import (
	"fmt"

	"github.com/coreos/coreos-cloudinit/system"
)

func SSHImportGithubUser(system_user string, github_user string) error {
	url := fmt.Sprintf("https://api.github.com/users/%s/keys", github_user)
	keys, err := fetchUserKeys(url)
	if err != nil {
		return err
	}

	key_name := fmt.Sprintf("github-%s", github_user)
	return system.AuthorizeSSHKeys(system_user, key_name, keys)
}
