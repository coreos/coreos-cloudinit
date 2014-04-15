package initialize

import (
	"testing"
)

func TestCloudConfigUsersBitbucketUser(t *testing.T) {

	contents := `
users:
  - name: ben
    coreos-ssh-import-bitbucket: benmccann
`
	cfg, err := NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error: %v", err)
	}

	if len(cfg.Users) != 1 {
		t.Fatalf("Parsed %d users, expected 1", cfg.Users)
	}

	user := cfg.Users[0]

	if user.Name != "ben" {
		t.Errorf("User name is %q, expected 'ben'", user.Name)
	}

	if user.SSHImportBitbucket != "benmccann" {
		t.Errorf("bitbucket user is %q, expected 'benmccann'", user.SSHImportBitbucketUser)
	}
}
