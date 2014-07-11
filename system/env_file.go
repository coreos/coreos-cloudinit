package system

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
)

type EnvFile struct {
	Vars map[string]string
	// mask File.Content, it shouldn't be used.
	Content interface{} `json:"-" yaml:"-"`
	*File
}

// only allow sh compatible identifiers
var validKey = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// match each line, optionally capturing valid identifiers, discarding dos line endings
var lineLexer = regexp.MustCompile(`(?m)^((?:([a-zA-Z0-9_]+)=)?.*?)\r?\n`)

// mergeEnvContents: Update the existing file contents with new values,
// preserving variable ordering and all content this code doesn't understand.
// All new values are appended to the bottom of the old.
func mergeEnvContents(old []byte, pending map[string]string) []byte {
	var buf bytes.Buffer
	var match [][]byte

	// it is awkward for the regex to handle a missing newline gracefully
	if len(old) != 0 && !bytes.HasSuffix(old, []byte{'\n'}) {
		old = append(old, byte('\n'))
	}

	for _, match = range lineLexer.FindAllSubmatch(old, -1) {
		key := string(match[2])
		if value, ok := pending[key]; ok {
			fmt.Fprintf(&buf, "%s=%s\n", key, value)
			delete(pending, key)
		} else {
			fmt.Fprintf(&buf, "%s\n", match[1])
		}
	}

	for key, value := range pending {
		fmt.Fprintf(&buf, "%s=%s\n", key, value)
	}

	return buf.Bytes()
}

// WriteEnvFile updates an existing env `KEY=value` formated file with
// new values provided in EnvFile.Vars; File.Content is ignored.
// Existing ordering and any unknown formatting such as comments are
// preserved. If no changes are required the file is untouched.
func WriteEnvFile(ef *EnvFile, root string) error {
	// validate new keys, mergeEnvContents uses pending to track writes
	pending := make(map[string]string, len(ef.Vars))
	for key, value := range ef.Vars {
		if !validKey.MatchString(key) {
			return fmt.Errorf("Invalid name %q for %s", key, ef.Path)
		}
		pending[key] = value
	}

	if len(pending) == 0 {
		return nil
	}

	oldContent, err := ioutil.ReadFile(path.Join(root, ef.Path))
	if err != nil {
		if os.IsNotExist(err) {
			oldContent = []byte{}
		} else {
			return err
		}
	}

	newContent := mergeEnvContents(oldContent, pending)
	if bytes.Equal(oldContent, newContent) {
		return nil
	}

	ef.File.Content = string(newContent)
	_, err = WriteFile(ef.File, root)
	return err
}
