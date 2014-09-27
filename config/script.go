package config

import (
	"strings"
)

type Script []byte

func IsScript(userdata string) bool {
	header := strings.SplitN(userdata, "\n", 2)[0]
	return strings.HasPrefix(header, "#!")
}

func NewScript(userdata string) (Script, error) {
	return Script(userdata), nil
}
