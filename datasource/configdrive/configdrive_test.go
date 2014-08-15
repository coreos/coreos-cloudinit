package configdrive

import (
	"os"
	"testing"
)

type mockFilesystem []string

func (m mockFilesystem) readFile(filename string) ([]byte, error) {
	for _, file := range m {
		if file == filename {
			return []byte(filename), nil
		}
	}
	return nil, os.ErrNotExist
}

func TestFetchMetadata(t *testing.T) {
	for _, tt := range []struct {
		root     string
		filename string
		files    mockFilesystem
	}{
		{
			"/",
			"",
			mockFilesystem{},
		},
		{
			"/",
			"/openstack/latest/meta_data.json",
			mockFilesystem([]string{"/openstack/latest/meta_data.json"}),
		},
		{
			"/media/configdrive",
			"/media/configdrive/openstack/latest/meta_data.json",
			mockFilesystem([]string{"/media/configdrive/openstack/latest/meta_data.json"}),
		},
	} {
		cd := configDrive{tt.root, tt.files.readFile}
		filename, err := cd.FetchMetadata()
		if err != nil {
			t.Fatalf("bad error for %q: want %q, got %q", tt, nil, err)
		}
		if string(filename) != tt.filename {
			t.Fatalf("bad path for %q: want %q, got %q", tt, tt.filename, filename)
		}
	}
}

func TestFetchUserdata(t *testing.T) {
	for _, tt := range []struct {
		root     string
		filename string
		files    mockFilesystem
	}{
		{
			"/",
			"",
			mockFilesystem{},
		},
		{
			"/",
			"/openstack/latest/user_data",
			mockFilesystem([]string{"/openstack/latest/user_data"}),
		},
		{
			"/media/configdrive",
			"/media/configdrive/openstack/latest/user_data",
			mockFilesystem([]string{"/media/configdrive/openstack/latest/user_data"}),
		},
	} {
		cd := configDrive{tt.root, tt.files.readFile}
		filename, err := cd.FetchUserdata()
		if err != nil {
			t.Fatalf("bad error for %q: want %q, got %q", tt, nil, err)
		}
		if string(filename) != tt.filename {
			t.Fatalf("bad path for %q: want %q, got %q", tt, tt.filename, filename)
		}
	}
}

func TestConfigRoot(t *testing.T) {
	for _, tt := range []struct {
		root       string
		configRoot string
	}{
		{
			"/",
			"/openstack",
		},
		{
			"/media/configdrive",
			"/media/configdrive/openstack",
		},
	} {
		cd := configDrive{tt.root, nil}
		if configRoot := cd.ConfigRoot(); configRoot != tt.configRoot {
			t.Fatalf("bad config root for %q: want %q, got %q", tt, tt.configRoot, configRoot)
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
