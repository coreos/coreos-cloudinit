package cloudinit

import (
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
)

// Add the provide SSH public key to the core user's list of
// authorized keys
func AuthorizeSSHKeys(user string, keysName string, keys []string) error {
	for i, key := range keys {
		keys[i] = strings.TrimSpace(key)
	}

	// join all keys with newlines, ensuring the resulting string
	// also ends with a newline
	joined := fmt.Sprintf("%s\n", strings.Join(keys, "\n"))

	cmd := exec.Command("update-ssh-keys", "-u", user, "-a", keysName)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		stdin.Close()
		return err
	}

	_, err = io.WriteString(stdin, joined)
	if err != nil {
		return err
	}

	stdin.Close()
	stdoutBytes, _ := ioutil.ReadAll(stdout)
	stderrBytes, _ := ioutil.ReadAll(stderr)

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("Call to update-ssh-keys failed with %v: %s %s", err, string(stdoutBytes), string(stderrBytes))
	}

	return nil
}
