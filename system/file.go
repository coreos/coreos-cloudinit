/*
   Copyright 2014 CoreOS, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package system

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/coreos/coreos-cloudinit/config"
)

// File is a top-level structure which embeds its underlying configuration,
// config.File, and provides the system-specific Permissions().
type File struct {
	config.File
}

func (f *File) Permissions() (os.FileMode, error) {
	if f.RawFilePermissions == "" {
		return os.FileMode(0644), nil
	}

	// Parse string representation of file mode as integer
	perm, err := strconv.ParseInt(f.RawFilePermissions, 0, 32)
	if err != nil {
		return 0, fmt.Errorf("Unable to parse file permissions %q as integer", f.RawFilePermissions)
	}
	return os.FileMode(perm), nil
}

func DecodeBase64Content(content string) ([]byte, error) {
	output, err := base64.StdEncoding.DecodeString(content)

	if err != nil {
		return nil, fmt.Errorf("Unable to decode base64: %v", err)
	}

	return output, nil
}

func DecodeGzipContent(content string) ([]byte, error) {
	gzr, err := gzip.NewReader(bytes.NewReader([]byte(content)))

	if err != nil {
		return nil, fmt.Errorf("Unable to decode gzip: %v", err)
	}
	defer gzr.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(gzr)

	return buf.Bytes(), nil
}

func DecodeContent(content string, encoding string) ([]byte, error) {
	switch encoding {
	case "":
		return []byte(content), nil

	case "b64", "base64":
		return DecodeBase64Content(content)

	case "gz", "gzip":
		return DecodeGzipContent(content)

	case "gz+base64", "gzip+base64", "gz+b64", "gzip+b64":
		gz, err := DecodeBase64Content(content)

		if err != nil {
			return nil, err
		}

		return DecodeGzipContent(string(gz))
	}

	return nil, fmt.Errorf("Unsupported encoding %s", encoding)

}

func WriteFile(f *File, root string) (string, error) {
	fullpath := path.Join(root, f.Path)
	dir := path.Dir(fullpath)
	log.Printf("Writing file to %q", fullpath)

	content, err := DecodeContent(f.Content, f.Encoding)

	if err != nil {
		return "", fmt.Errorf("Unable to decode %s (%v)", f.Path, err)
	}

	if err := EnsureDirectoryExists(dir); err != nil {
		return "", err
	}

	perm, err := f.Permissions()
	if err != nil {
		return "", err
	}

	var tmp *os.File
	// Create a temporary file in the same directory to ensure it's on the same filesystem
	if tmp, err = ioutil.TempFile(dir, "cloudinit-temp"); err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(tmp.Name(), content, perm); err != nil {
		return "", err
	}

	if err := tmp.Close(); err != nil {
		return "", err
	}

	// Ensure the permissions are as requested (since WriteFile can be affected by sticky bit)
	if err := os.Chmod(tmp.Name(), perm); err != nil {
		return "", err
	}

	if f.Owner != "" {
		// We shell out since we don't have a way to look up unix groups natively
		cmd := exec.Command("chown", f.Owner, tmp.Name())
		if err := cmd.Run(); err != nil {
			return "", err
		}
	}

	if err := os.Rename(tmp.Name(), fullpath); err != nil {
		return "", err
	}

	log.Printf("Wrote file to %q", fullpath)
	return fullpath, nil
}

func EnsureDirectoryExists(dir string) error {
	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", dir)
		}
	} else {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
