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

func TestCDFetchMetadata(t *testing.T) {
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
			"/ec2/2009-04-04/meta-data.json",
			mockFilesystem([]string{"/ec2/2009-04-04/meta-data.json"}),
		},
		{
			"/media/configdrive",
			"/media/configdrive/ec2/2009-04-04/meta-data.json",
			mockFilesystem([]string{"/media/configdrive/ec2/2009-04-04/meta-data.json"}),
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

func TestCDFetchUserdata(t *testing.T) {
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
			"/ec2/2009-04-04/user-data",
			mockFilesystem([]string{"/ec2/2009-04-04/user-data"}),
		},
		{
			"/",
			"/openstack/latest/user_data",
			mockFilesystem([]string{"/openstack/latest/user_data"}),
		},
		{
			"/",
			"/ec2/2009-04-04/user-data",
			mockFilesystem([]string{"/openstack/latest/user_data", "/ec2/2009-04-04/user-data"}),
		},
		{
			"/media/configdrive",
			"/media/configdrive/ec2/2009-04-04/user-data",
			mockFilesystem([]string{"/media/configdrive/ec2/2009-04-04/user-data"}),
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

func TestCDConfigRoot(t *testing.T) {
	for _, tt := range []struct {
		root       string
		configRoot string
	}{
		{
			"/",
			"/openstack/latest",
		},
		{
			"/media/configdrive",
			"/media/configdrive/openstack/latest",
		},
	} {
		cd := configDrive{tt.root, nil}
		if configRoot := cd.ConfigRoot(); configRoot != tt.configRoot {
			t.Fatalf("bad config root for %q: want %q, got %q", tt, tt.configRoot, configRoot)
		}
	}
}
