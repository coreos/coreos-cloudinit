package config

type File struct {
	Encoding           string `yaml:"-"`
	Content            string `yaml:"content"`
	Owner              string `yaml:"owner"`
	Path               string `yaml:"path"`
	RawFilePermissions string `yaml:"permissions"`
}
