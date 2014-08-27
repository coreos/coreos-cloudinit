package ec2

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/datasource/metadata"
	"github.com/coreos/coreos-cloudinit/datasource/metadata/test"
	"github.com/coreos/coreos-cloudinit/pkg"
)

func TestType(t *testing.T) {
	want := "ec2-metadata-service"
	if kind := (metadataService{}).Type(); kind != want {
		t.Fatalf("bad type: want %q, got %q", want, kind)
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
			err: fmt.Errorf("test error"),
			tests: []struct {
				path string
				val  []string
			}{
				{"", nil},
			},
		},
	} {
		service := metadataService{metadata.MetadataService{
			Client: &test.HttpClient{s.resources, s.err},
		}}
		for _, tt := range s.tests {
			attrs, err := service.fetchAttributes(tt.path)
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
			err: fmt.Errorf("test error"),
			tests: []struct {
				path string
				val  string
			}{
				{"", ""},
			},
		},
	} {
		service := metadataService{metadata.MetadataService{
			Client: &test.HttpClient{s.resources, s.err},
		}}
		for _, tt := range s.tests {
			attr, err := service.fetchAttribute(tt.path)
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
		root         string
		metadataPath string
		resources    map[string]string
		expect       []byte
		clientErr    error
		expectErr    error
	}{
		{
			root:         "/",
			metadataPath: "2009-04-04/meta-data",
			resources: map[string]string{
				"/2009-04-04/meta-data/public-keys": "bad\n",
			},
			expectErr: fmt.Errorf("malformed public key: \"bad\""),
		},
		{
			root:         "/",
			metadataPath: "2009-04-04/meta-data",
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
		service := &metadataService{metadata.MetadataService{
			Root:         tt.root,
			Client:       &test.HttpClient{tt.resources, tt.clientErr},
			MetadataPath: tt.metadataPath,
		}}
		metadata, err := service.FetchMetadata()
		if Error(err) != Error(tt.expectErr) {
			t.Fatalf("bad error (%q): want %q, got %q", tt.resources, tt.expectErr, err)
		}
		if !bytes.Equal(metadata, tt.expect) {
			t.Fatalf("bad fetch (%q): want %q, got %q", tt.resources, tt.expect, metadata)
		}
	}
}

func Error(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
