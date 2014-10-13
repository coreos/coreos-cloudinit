package initialize

import (
	"testing"
)

func TestCloudConfigUsersGithubUser(t *testing.T) {

	contents := `
users:
  - name: elroy
    coreos-ssh-import-github: bcwaldon
`
	cfg, err := NewCloudConfig(contents)
	if err != nil {
		t.Fatalf("Encountered unexpected error: %v", err)
	}

	if len(cfg.Users) != 1 {
		t.Fatalf("Parsed %d users, expected 1", cfg.Users)
	}

	user := cfg.Users[0]

	if user.Name != "elroy" {
		t.Errorf("User name is %q, expected 'elroy'", user.Name)
	}

	if user.SSHImportGithubUser != "bcwaldon" {
		t.Errorf("github user is %q, expected 'bcwaldon'", user.SSHImportGithubUser)
	}
}
