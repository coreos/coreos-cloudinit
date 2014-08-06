package ec2

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/pkg"
)

type testHttpClient struct {
	resources map[string]string
	err       error
}

func (t *testHttpClient) GetRetry(url string) ([]byte, error) {
	if t.err != nil {
		return nil, t.err
	}
	if val, ok := t.resources[url]; ok {
		return []byte(val), nil
	} else {
		return nil, pkg.ErrNotFound{fmt.Errorf("not found: %q", url)}
	}
}

func (t *testHttpClient) Get(url string) ([]byte, error) {
	return t.GetRetry(url)
}

func TestAvailabilityChanges(t *testing.T) {
	want := true
	if ac := (metadataService{}).AvailabilityChanges(); ac != want {
		t.Fatalf("bad AvailabilityChanges: want %q, got %q", want, ac)
	}
}

func TestType(t *testing.T) {
	want := "ec2-metadata-service"
	if kind := (metadataService{}).Type(); kind != want {
		t.Fatalf("bad type: want %q, got %q", want, kind)
	}
}

func TestIsAvailable(t *testing.T) {
	for _, tt := range []struct {
		root      string
		resources map[string]string
		expect    bool
	}{
		{
			root: "/",
			resources: map[string]string{
				"/2009-04-04": "",
			},
			expect: true,
		},
		{
			root:      "/",
			resources: map[string]string{},
			expect:    false,
		},
	} {
		service := &metadataService{tt.root, &testHttpClient{tt.resources, nil}}
		if a := service.IsAvailable(); a != tt.expect {
			t.Fatalf("bad isAvailable (%q): want %q, got %q", tt.resources, tt.expect, a)
		}
	}
}

func TestFetchUserdata(t *testing.T) {
	for _, tt := range []struct {
		root      string
		resources map[string]string
		userdata  []byte
		clientErr error
		expectErr error
	}{
		{
			root: "/",
			resources: map[string]string{
				"/2009-04-04/user-data": "hello",
			},
			userdata: []byte("hello"),
		},
		{
			root:      "/",
			clientErr: pkg.ErrNotFound{fmt.Errorf("test not found error")},
			userdata:  []byte{},
		},
		{
			root:      "/",
			clientErr: pkg.ErrTimeout{fmt.Errorf("test timeout error")},
			expectErr: pkg.ErrTimeout{fmt.Errorf("test timeout error")},
		},
	} {
		service := &metadataService{tt.root, &testHttpClient{tt.resources, tt.clientErr}}
		data, err := service.FetchUserdata()
		if Error(err) != Error(tt.expectErr) {
			t.Fatalf("bad error (%q): want %q, got %q", tt.resources, tt.expectErr, err)
		}
		if !bytes.Equal(data, tt.userdata) {
			t.Fatalf("bad userdata (%q): want %q, got %q", tt.resources, tt.userdata, data)
		}
	}
}

func TestUrls(t *testing.T) {
	for _, tt := range []struct {
		root       string
		expectRoot string
		userdata   string
		metadata   string
	}{
		{
			root:       "/",
			expectRoot: "/",
			userdata:   "/2009-04-04/user-data",
			metadata:   "/2009-04-04/meta-data",
		},
		{
			root:       "http://169.254.169.254/",
			expectRoot: "http://169.254.169.254/",
			userdata:   "http://169.254.169.254/2009-04-04/user-data",
			metadata:   "http://169.254.169.254/2009-04-04/meta-data",
		},
	} {
		service := &metadataService{tt.root, nil}
		if url := service.userdataUrl(); url != tt.userdata {
			t.Fatalf("bad url (%q): want %q, got %q", tt.root, tt.userdata, url)
		}
		if url := service.metadataUrl(); url != tt.metadata {
			t.Fatalf("bad url (%q): want %q, got %q", tt.root, tt.metadata, url)
		}
		if url := service.ConfigRoot(); url != tt.expectRoot {
			t.Fatalf("bad url (%q): want %q, got %q", tt.root, tt.expectRoot, url)
		}
	}
}

func TestFetchAttributes(t *testing.T) {
	for _, s := range []struct {
		resources map[string]string
		err       error
		tests     []struct {
			path string
			val  []string
		}
	}{
		{
			resources: map[string]string{
				"/":      "a\nb\nc/",
				"/c/":    "d\ne/",
				"/c/e/":  "f",
				"/a":     "1",
				"/b":     "2",
				"/c/d":   "3",
				"/c/e/f": "4",
			},
			tests: []struct {
				path string
				val  []string
			}{
				{"/", []string{"a", "b", "c/"}},
				{"/b", []string{"2"}},
				{"/c/d", []string{"3"}},
				{"/c/e/", []string{"f"}},
			},
		},
		{
			err: pkg.ErrNotFound{fmt.Errorf("test error")},
			tests: []struct {
				path string
				val  []string
			}{
				{"", nil},
			},
		},
	} {
		client := &testHttpClient{s.resources, s.err}
		for _, tt := range s.tests {
			attrs, err := fetchAttributes(client, tt.path)
			if err != s.err {
				t.Fatalf("bad error for %q (%q): want %q, got %q", tt.path, s.resources, s.err, err)
			}
			if !reflect.DeepEqual(attrs, tt.val) {
				t.Fatalf("bad fetch for %q (%q): want %q, got %q", tt.path, s.resources, tt.val, attrs)
			}
		}
	}
}

func TestFetchAttribute(t *testing.T) {
	for _, s := range []struct {
		resources map[string]string
		err       error
		tests     []struct {
			path string
			val  string
		}
	}{
		{
			resources: map[string]string{
				"/":      "a\nb\nc/",
				"/c/":    "d\ne/",
				"/c/e/":  "f",
				"/a":     "1",
				"/b":     "2",
				"/c/d":   "3",
				"/c/e/f": "4",
			},
			tests: []struct {
				path string
				val  string
			}{
				{"/a", "1"},
				{"/b", "2"},
				{"/c/d", "3"},
				{"/c/e/f", "4"},
			},
		},
		{
			err: pkg.ErrNotFound{fmt.Errorf("test error")},
			tests: []struct {
				path string
				val  string
			}{
				{"", ""},
			},
		},
	} {
		client := &testHttpClient{s.resources, s.err}
		for _, tt := range s.tests {
			attr, err := fetchAttribute(client, tt.path)
			if err != s.err {
				t.Fatalf("bad error for %q (%q): want %q, got %q", tt.path, s.resources, s.err, err)
			}
			if attr != tt.val {
				t.Fatalf("bad fetch for %q (%q): want %q, got %q", tt.path, s.resources, tt.val, attr)
			}
		}
	}
}

func TestFetchMetadata(t *testing.T) {
	for _, tt := range []struct {
		root      string
		resources map[string]string
		expect    []byte
		clientErr error
		expectErr error
	}{
		{
			root: "/",
			resources: map[string]string{
				"/2009-04-04/meta-data/public-keys": "bad\n",
			},
			expectErr: fmt.Errorf("malformed public key: \"bad\""),
		},
		{
			root: "/",
			resources: map[string]string{
				"/2009-04-04/meta-data/hostname":                    "host",
				"/2009-04-04/meta-data/local-ipv4":                  "1.2.3.4",
				"/2009-04-04/meta-data/public-ipv4":                 "5.6.7.8",
				"/2009-04-04/meta-data/public-keys":                 "0=test1\n",
				"/2009-04-04/meta-data/public-keys/0":               "openssh-key",
				"/2009-04-04/meta-data/public-keys/0/openssh-key":   "key",
				"/2009-04-04/meta-data/network_config/content_path": "path",
			},
			expect: []byte(`{"hostname":"host","local-ipv4":"1.2.3.4","network_config":{"content_path":"path"},"public-ipv4":"5.6.7.8","public_keys":{"test1":"key"}}`),
		},
		{
			clientErr: pkg.ErrTimeout{fmt.Errorf("test error")},
			expectErr: pkg.ErrTimeout{fmt.Errorf("test error")},
		},
	} {
		service := &metadataService{tt.root, &testHttpClient{tt.resources, tt.clientErr}}
		metadata, err := service.FetchMetadata()
		if Error(err) != Error(tt.expectErr) {
			t.Fatalf("bad error (%q): want %q, got %q", tt.resources, tt.expectErr, err)
		}
		if !bytes.Equal(metadata, tt.expect) {
			t.Fatalf("bad fetch (%q): want %q, got %q", tt.resources, tt.expect, metadata)
		}
	}
}

func TestNewDatasource(t *testing.T) {
	for _, tt := range []struct {
		root       string
		expectRoot string
	}{
		{
			root:       "",
			expectRoot: "/",
		},
		{
			root:       "/",
			expectRoot: "/",
		},
		{
			root:       "http://169.254.169.254",
			expectRoot: "http://169.254.169.254/",
		},
		{
			root:       "http://169.254.169.254/",
			expectRoot: "http://169.254.169.254/",
		},
	} {
		service := NewDatasource(tt.root)
		if service.root != tt.expectRoot {
			t.Fatalf("bad root (%q): want %q, got %q", tt.root, tt.expectRoot, service.root)
		}
	}
}

func Error(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
