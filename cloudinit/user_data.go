package cloudinit

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strings"
)

func ParseUserData(contents []byte) (interface{}, error) {
	bytereader := bytes.NewReader(contents)
	bufreader := bufio.NewReader(bytereader)
	header, _ := bufreader.ReadString('\n')

	if strings.HasPrefix(header, "#!") {
		log.Printf("Parsing user-data as script")
		return Script(contents), nil

	} else if header == "#cloud-config\n" {
		log.Printf("Parsing user-data as cloud-config")
		cfg, err := NewCloudConfig(contents)
		if err != nil {
			log.Fatal(err.Error())
		}
		return *cfg, nil
	} else {
		return nil, fmt.Errorf("Unrecognized user-data header: %s", header)
	}
}
