package initialize

import (
	"fmt"

	"github.com/coreos/coreos-cloudinit/system"
)

func SSHImportBitbucketUser(system_user string, bitbucket_user string) error {
	url := fmt.Sprintf("https://bitbucket.org/api/1.0/users/%s/ssh-keys", bitbucket_user)
	keys, err := fetchUserKeys(url)
	if err != nil {
		return err
	}

	key_name := fmt.Sprintf("bitbucket-%s", bitbucket_user)
	return system.AuthorizeSSHKeys(system_user, key_name, keys)
}
