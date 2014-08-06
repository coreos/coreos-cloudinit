package url

import "github.com/coreos/coreos-cloudinit/pkg"

type remoteFile struct {
	url string
}

func NewDatasource(url string) *remoteFile {
	return &remoteFile{url}
}

func (f *remoteFile) IsAvailable() bool {
	client := pkg.NewHttpClient()
	_, err := client.Get(f.url)
	return (err == nil)
}

func (f *remoteFile) AvailabilityChanges() bool {
	return true
}

func (f *remoteFile) ConfigRoot() string {
	return ""
}

func (f *remoteFile) FetchMetadata() ([]byte, error) {
	return []byte{}, nil
}

func (f *remoteFile) FetchUserdata() ([]byte, error) {
	client := pkg.NewHttpClient()
	return client.GetRetry(f.url)
}

func (f *remoteFile) Type() string {
	return "url"
}
