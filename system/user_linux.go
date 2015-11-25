// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package system

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"
)

func UserExists(u *config.User) bool {
       return exec.Command("getent", "passwd", u.Name).Run() == nil
}

func CreateUser(u *config.User) error {
	args := []string{}

	if u.PasswordHash != "" {
		args = append(args, "--password", u.PasswordHash)
	} else {
		args = append(args, "--password", "*")
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
		args = append(args, "--gid", u.PrimaryGroup)
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

	if u.Shell != "" {
		args = append(args, "--shell", u.Shell)
	}

	args = append(args, u.Name)

	output, err := exec.Command("useradd", args...).CombinedOutput()
	if err != nil {
		log.Printf("Command 'useradd %s' failed: %v\n%s", strings.Join(args, " "), err, output)
		return err
	}

	return nil
}

func IsLockedUser(u *config.User) bool {
	output, err := exec.Command("getent", "shadow", u.Name).CombinedOutput()
	if err == nil {
		fields := strings.Split(string(output), ":")
		if len(fields[1]) > 1 && fields[1][0] == '!' {
			return true
		}
	}
	return false
}

func LockUnlockUser(u *config.User) error {
	args := []string{}

	if u.LockPasswd {
		args = append(args, "-l")
	} else {
		if !IsLockedUser(u) {
			return nil
		}
		args = append(args, "-u")
	}

	args = append(args, u.Name)

	output, err := exec.Command("passwd", args...).CombinedOutput()
	if err != nil {
		log.Printf("Command 'passwd %s' failed: %v\n%s", strings.Join(args, " "), err, output)
	}
	return err
}

func SetUserPassword(user, hash string) error {
	cmd := exec.Command("chpasswd", "-e")

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

func UserHome(name string) (string, error) {
	output, err := exec.Command("getent", "passwd", name).CombinedOutput()
	if err != nil {
		return "", err
	}
	passwd := strings.Split(string(output), ":")
	return passwd[5], nil
}
