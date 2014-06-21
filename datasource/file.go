package datasource

import (
	"io/ioutil"
)

type localFile struct {
	path string
}

func NewLocalFile(path string) *localFile {
	return &localFile{path}
}

func (self *localFile) ConfigRoot() string {
	return ""
}

func (self *localFile) FetchMetadata() ([]byte, error) {
	return []byte{}, nil
}

func (self *localFile) FetchUserdata() ([]byte, error) {
	return ioutil.ReadFile(self.path)
}

func (self *localFile) Type() string {
	return "local-file"
}
