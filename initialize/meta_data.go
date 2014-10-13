package initialize

import (
	"encoding/json"
	"sort"
)

// ParseMetaData parses a JSON blob in the OpenStack metadata service format,
// and converts it to a partially hydrated CloudConfig.
func ParseMetaData(contents string) (*CloudConfig, error) {
	if len(contents) == 0 {
		return nil, nil
	}
	var metadata struct {
		SSHAuthorizedKeyMap map[string]string `json:"public_keys"`
		Hostname            string            `json:"hostname"`
		NetworkConfig       struct {
			ContentPath string `json:"content_path"`
		} `json:"network_config"`
	}
	if err := json.Unmarshal([]byte(contents), &metadata); err != nil {
		return nil, err
	}

	var cfg CloudConfig
	if len(metadata.SSHAuthorizedKeyMap) > 0 {
		cfg.SSHAuthorizedKeys = make([]string, 0, len(metadata.SSHAuthorizedKeyMap))
		for _, name := range sortedKeys(metadata.SSHAuthorizedKeyMap) {
			cfg.SSHAuthorizedKeys = append(cfg.SSHAuthorizedKeys, metadata.SSHAuthorizedKeyMap[name])
		}
	}
	cfg.Hostname = metadata.Hostname
	cfg.NetworkConfigPath = metadata.NetworkConfig.ContentPath
	return &cfg, nil
}

// ExtractIPsFromMetaData parses a JSON blob in the OpenStack metadata service
// format and returns a substitution map possibly containing private_ipv4,
// public_ipv4, private_ipv6, and public_ipv6 addresses.
func ExtractIPsFromMetadata(contents []byte) (map[string]string, error) {
	var ips struct {
		PublicIPv4  string `json:"public-ipv4"`
		PrivateIPv4 string `json:"local-ipv4"`
		PublicIPv6  string `json:"public-ipv6"`
		PrivateIPv6 string `json:"local-ipv6"`
	}
	if err := json.Unmarshal(contents, &ips); err != nil {
		return nil, err
	}
	m := make(map[string]string)
	if ips.PrivateIPv4 != "" {
		m["$private_ipv4"] = ips.PrivateIPv4
	}
	if ips.PublicIPv4 != "" {
		m["$public_ipv4"] = ips.PublicIPv4
	}
	if ips.PrivateIPv6 != "" {
		m["$private_ipv6"] = ips.PrivateIPv6
	}
	if ips.PublicIPv6 != "" {
		m["$public_ipv6"] = ips.PublicIPv6
	}
	return m, nil
}

func sortedKeys(m map[string]string) (keys []string) {
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return
}
