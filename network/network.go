package network

import (
	"strings"
)

func ProcessDebianNetconf(config string) ([]InterfaceGenerator, error) {
	lines := formatConfig(config)
	stanzas, err := parseStanzas(lines)
	if err != nil {
		return nil, err
	}

	interfaces := make([]*stanzaInterface, 0, len(stanzas))
	for _, stanza := range stanzas {
		switch s := stanza.(type) {
		case *stanzaInterface:
			interfaces = append(interfaces, s)
		}
	}

	return buildInterfaces(interfaces), nil
}

func formatConfig(config string) []string {
	lines := []string{}
	config = strings.Replace(config, "\\\n", "", -1)
	for config != "" {
		split := strings.SplitN(config, "\n", 2)
		line := strings.TrimSpace(split[0])

		if len(split) == 2 {
			config = split[1]
		} else {
			config = ""
		}

		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		lines = append(lines, line)
	}
	return lines
}
