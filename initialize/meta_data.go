package initialize

import (
	"encoding/json"
)

func ParseMetaData(contents string) (cfg CloudConfig, err error) {
	var metadata struct {
		SSHAuthorizedKeyMap map[string]string `json:"public_keys"`
		Hostname            string            `json:"hostname"`
		NetworkConfig       struct {
			ContentPath string `json:"content_path"`
		} `json:"network_config"`
	}
	if err = json.Unmarshal([]byte(contents), &metadata); err != nil {
		return
	}

	cfg.SSHAuthorizedKeys = make([]string, 0, len(metadata.SSHAuthorizedKeyMap))
	for _, key := range metadata.SSHAuthorizedKeyMap {
		cfg.SSHAuthorizedKeys = append(cfg.SSHAuthorizedKeys, key)
	}
	cfg.Hostname = metadata.Hostname
	cfg.NetworkConfigPath = metadata.NetworkConfig.ContentPath
	return
}
