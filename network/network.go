package network

import (
	"fmt"
	"io"
	"os"
	"path"
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

func WriteConfigs(configPath string, interfaces []InterfaceGenerator) error {
	if err := os.MkdirAll(configPath, os.ModePerm+os.ModeDir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, iface := range interfaces {
		filename := path.Join(configPath, fmt.Sprintf("%s.netdev", iface.Name()))
		if err := writeConfig(filename, iface.GenerateNetdevConfig()); err != nil {
			return err
		}
		filename = path.Join(configPath, fmt.Sprintf("%s.link", iface.Name()))
		if err := writeConfig(filename, iface.GenerateLinkConfig()); err != nil {
			return err
		}
		filename = path.Join(configPath, fmt.Sprintf("%s.network", iface.Name()))
		if err := writeConfig(filename, iface.GenerateNetworkConfig()); err != nil {
			return err
		}
	}
	return nil
}

func writeConfig(filename string, config string) error {
	if config == "" {
		return nil
	}

	if file, err := os.Create(filename); err == nil {
		io.WriteString(file, config)
		file.Close()
		return nil
	} else {
		return err
	}
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
