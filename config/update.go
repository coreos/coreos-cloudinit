package config

type Update struct {
	RebootStrategy string `yaml:"reboot-strategy" env:"REBOOT_STRATEGY" valid:"best-effort,etcd-lock,reboot,off"`
	Group          string `yaml:"group"           env:"GROUP"`
	Server         string `yaml:"server"          env:"SERVER"`
}
