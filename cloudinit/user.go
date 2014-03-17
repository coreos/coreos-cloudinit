package cloudinit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"os/user"
	"strings"
)

type User struct {
	Name                string   `yaml:"name"`
	PasswordHash        string   `yaml:"passwd"`
	SSHAuthorizedKeys   []string `yaml:"ssh-authorized-keys"`
	SSHImportGithubUser string   `yaml:"ssh-import-github"`
	GECOS               string   `yaml:"gecos"`
	Homedir             string   `yaml:"homedir"`
	NoCreateHome        bool     `yaml:"no-create-home"`
	PrimaryGroup        string   `yaml:"primary-group"`
	Groups              []string `yaml:"groups"`
	NoUserGroup         bool     `yaml:"no-user-group"`
	System              bool     `yaml:"system"`
	NoLogInit           bool     `yaml:"no-log-init"`
}

func UserExists(u *User) bool {
	_, err := user.Lookup(u.Name)
	return err == nil
}

func CreateUser(u *User) error {
	args := []string{}

	if u.PasswordHash != "" {
		args = append(args, "--password", u.PasswordHash)
	}

	if u.GECOS != "" {
		args = append(args, "--comment", fmt.Sprintf("%q", u.GECOS))
	}

	if u.Homedir != "" {
		args = append(args, "--home-dir", u.Homedir)
	}

	if u.NoCreateHome {
		args = append(args, "--no-create-home")
	} else {
		args = append(args, "--create-home")
	}

	if u.PrimaryGroup != "" {
		args = append(args, "--primary-group", u.PrimaryGroup)
	}

	if len(u.Groups) > 0 {
		args = append(args, "--groups", strings.Join(u.Groups, ","))
	}

	if u.NoUserGroup {
		args = append(args, "--no-user-group")
	}

	if u.System {
		args = append(args, "--system")
	}

	if u.NoLogInit {
		args = append(args, "--no-log-init")
	}

	args = append(args, u.Name)

	output, err := exec.Command("useradd", args...).CombinedOutput()
	if err != nil {
		log.Printf("Command 'useradd %s' failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return err
}

func SetUserPassword(user, hash string) error {
	cmd := exec.Command("/usr/sbin/chpasswd", "-e")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	arg := fmt.Sprintf("%s:%s", user, hash)
	_, err = stdin.Write([]byte(arg))
	if err != nil {
		return err
	}
	stdin.Close()

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

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
	err = AuthorizeSSHKeys(system_user, key_name, keys)
	if err != nil {
		return err
	}
	return nil
}
