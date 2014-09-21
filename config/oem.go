package config

type OEM struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	VersionID    string `yaml:"version-id"`
	HomeURL      string `yaml:"home-url"`
	BugReportURL string `yaml:"bug-report-url"`
}
