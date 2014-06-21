package datasource

import "github.com/coreos/coreos-cloudinit/pkg"

type remoteFile struct {
	url string
}

func NewRemoteFile(url string) *remoteFile {
	return &remoteFile{url}
}

func (f *remoteFile) ConfigRoot() string {
	return ""
}

func (f *remoteFile) FetchMetadata() ([]byte, error) {
	return []byte{}, nil
}

func (f *remoteFile) FetchUserdata() ([]byte, error) {
	client := pkg.NewHttpClient()
	return client.Get(f.url)
}

func (f *remoteFile) Type() string {
	return "url"
}
