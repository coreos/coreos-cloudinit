package initialize

import "encoding/json"

// ParseMetaData parses a JSON blob in the OpenStack metadata service format, and
// converts it to a CloudConfig
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

// ExtractIPsFromMetaData parses a JSON blob in the OpenStack metadata service format,
// and returns a substitution map possibly containing private_ipv4 and public_ipv4 addresses
func ExtractIPsFromMetadata(contents []byte) (map[string]string, error) {
	var ips struct {
		Public  string `json:"public-ipv4"`
		Private string `json:"local-ipv4"`
	}
	if err := json.Unmarshal(contents, &ips); err != nil {
		return nil, err
	}
	m := make(map[string]string)
	if ips.Private != "" {
		m["$private_ipv4"] = ips.Private
	}
	if ips.Public != "" {
		m["$public_ipv4"] = ips.Public
	}
	return m, nil
}
