package opennebula_drive

import (
	"testing"
	"bytes"
)

func TestFetchMetadata(t *testing.T) {
	for _, tt := range []struct {
		path, expectValue string
	} {
		{"testdata/with_data/", "12345"},
		{"testdata/without_data/", ""},
	} {
		ds := NewDatasource(tt.path)
		metaData, err := ds.FetchMetadata()
		
		if err != nil {
			t.Fatalf("Error during using fetchUserdata.\n")
		}
		if !(metaData.SSHPublicKeys["SSH_KEY"] == tt.expectValue) {
			t.Fatalf("bad userData: want %q, got %q", tt.expectValue, metaData.SSHPublicKeys["SSH_KEY"])
		}
	}
}


func TestFetchUserdata(t *testing.T) {
	for _, tt := range []struct {
		path string
		expectValue []byte
	} {
		{"testdata/with_data/", []byte("123456")},
		{"testdata/without_data/", []byte("")},
	} {
		ds := NewDatasource(tt.path)
		userData, err := ds.FetchUserdata()
		
		if err != nil {
			t.Fatalf("Error during using fetchUserdata.\n")
		}
		if !bytes.Equal(userData, tt.expectValue) {
			t.Fatalf("bad userData: want %q, got %q", tt.expectValue, userData)
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
			expectRoot: "",
		},
		{
			root:       "/media/configdrive",
			expectRoot: "/media/configdrive",
		},
	} {
		service := NewDatasource(tt.root)
		if service.root != tt.expectRoot {
			t.Fatalf("bad root (%q): want %q, got %q", tt.root, tt.expectRoot, service.root)
		}
	}
}

func Testdecodebase64(t *testing.T) {
	for _, tt := range []struct {
		text, expectText string
	} {
		{"hello world", "aGVsbG8gd29ybGQ="},
	} {
		decodedText, err := decodeBase64(tt.text)
		
		if err != nil {
			t.Fatalf("Error during decoding from base64.\n")
		}
		if !(decodedText == tt.expectText) {
			t.Fatalf("bad userData: want %q, got %q", tt.expectText, decodedText)
		}
	}
}