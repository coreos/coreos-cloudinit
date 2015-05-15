package rpcvmx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/coreos-cloudinit/Godeps/_workspace/src/github.com/sigma/vmw-guestinfo/rpcout"
)

// Config gives access to the vmx config through the VMware backdoor
type Config struct{}

// NewConfig creates a new Config object
func NewConfig() *Config {
	return &Config{}
}

// GetString returns the config string in the guestinfo.* namespace
func (c *Config) GetString(key string, defaultValue string) (string, error) {
	out, ok, err := rpcout.SendOne("info-get guestinfo.%s", key)
	if err != nil {
		return "", err
	} else if !ok {
		return defaultValue, nil
	}
	return string(out), nil
}

// GetBool returns the config boolean in the guestinfo.* namespace
func (c *Config) GetBool(key string, defaultValue bool) (bool, error) {
	val, err := c.GetString(key, fmt.Sprintf("%t", defaultValue))
	if err != nil {
		return false, err
	}
	switch strings.ToLower(val) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return defaultValue, nil
	}
}

// GetInt returns the config integer in the guestinfo.* namespace
func (c *Config) GetInt(key string, defaultValue int) (int, error) {
	val, err := c.GetString(key, "")
	if err != nil {
		return 0, err
	}
	res, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue, nil
	}
	return res, nil
}
