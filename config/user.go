package config

type User struct {
	Name                string   `yaml:"name"`
	PasswordHash        string   `yaml:"passwd"`
	SSHAuthorizedKeys   []string `yaml:"ssh-authorized-keys"`
	SSHImportGithubUser string   `yaml:"coreos-ssh-import-github"`
	SSHImportURL        string   `yaml:"coreos-ssh-import-url"`
	GECOS               string   `yaml:"gecos"`
	Homedir             string   `yaml:"homedir"`
	NoCreateHome        bool     `yaml:"no-create-home"`
	PrimaryGroup        string   `yaml:"primary-group"`
	Groups              []string `yaml:"groups"`
	NoUserGroup         bool     `yaml:"no-user-group"`
	System              bool     `yaml:"system"`
	NoLogInit           bool     `yaml:"no-log-init"`
}
