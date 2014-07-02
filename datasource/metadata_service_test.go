package datasource

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/pkg"
)

type TestHttpClient struct {
	metadata map[string]string
	err      error
}

func (t *TestHttpClient) GetRetry(url string) ([]byte, error) {
	if t.err != nil {
		return nil, t.err
	}
	if val, ok := t.metadata[url]; ok {
		return []byte(val), nil
	} else {
		return nil, pkg.ErrNotFound{fmt.Errorf("not found: %q", url)}
	}
}

func TestFetchAttributes(t *testing.T) {
	for _, s := range []struct {
		metadata map[string]string
		err      error
		tests    []struct {
			path string
			val  []string
		}
	}{
		{
			metadata: map[string]string{
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
		client := &TestHttpClient{s.metadata, s.err}
		for _, tt := range s.tests {
			attrs, err := fetchAttributes(client, tt.path)
			if err != s.err {
				t.Fatalf("bad error for %q (%q): want %q, got %q", tt.path, s.metadata, s.err, err)
			}
			if !reflect.DeepEqual(attrs, tt.val) {
				t.Fatalf("bad fetch for %q (%q): want %q, got %q", tt.path, s.metadata, tt.val, attrs)
			}
		}
	}
}

func TestFetchAttribute(t *testing.T) {
	for _, s := range []struct {
		metadata map[string]string
		err      error
		tests    []struct {
			path string
			val  string
		}
	}{
		{
			metadata: map[string]string{
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
		client := &TestHttpClient{s.metadata, s.err}
		for _, tt := range s.tests {
			attr, err := fetchAttribute(client, tt.path)
			if err != s.err {
				t.Fatalf("bad error for %q (%q): want %q, got %q", tt.path, s.metadata, s.err, err)
			}
			if attr != tt.val {
				t.Fatalf("bad fetch for %q (%q): want %q, got %q", tt.path, s.metadata, tt.val, attr)
			}
		}
	}
}

func TestFetchMetadata(t *testing.T) {
	for _, tt := range []struct {
		metadata map[string]string
		err      error
		expect   []byte
	}{
		{
			metadata: map[string]string{
				"http://169.254.169.254/2009-04-04/meta-data/hostname":                    "host",
				"http://169.254.169.254/2009-04-04/meta-data/public-keys":                 "0=test1\n",
				"http://169.254.169.254/2009-04-04/meta-data/public-keys/0":               "openssh-key",
				"http://169.254.169.254/2009-04-04/meta-data/public-keys/0/openssh-key":   "key",
				"http://169.254.169.254/2009-04-04/meta-data/network_config/content_path": "path",
			},
			expect: []byte(`{"hostname":"host","network_config":{"content_path":"path"},"public_keys":{"test1":"key"}}`),
		},
		{err: pkg.ErrTimeout{fmt.Errorf("test error")}},
	} {
		client := &TestHttpClient{tt.metadata, tt.err}
		metadata, err := fetchMetadata(client)
		if err != tt.err {
			t.Fatalf("bad error (%q): want %q, got %q", tt.metadata, tt.err, err)
		}
		if !bytes.Equal(metadata, tt.expect) {
			t.Fatalf("bad fetch (%q): want %q, got %q", tt.metadata, tt.expect, metadata)
		}
	}
}
