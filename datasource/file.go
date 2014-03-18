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

func (self *localFile) Fetch() ([]byte, error) {
	return ioutil.ReadFile(self.path)
}

func (self *localFile) Type() string {
	return "local-file"
}
