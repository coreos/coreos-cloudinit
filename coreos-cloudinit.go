package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/initialize"
	"github.com/coreos/coreos-cloudinit/system"
)

const version = "0.2.0+git"

func main() {
	var printVersion bool
	flag.BoolVar(&printVersion, "version", false, "Print the version and exit")

	var ignoreFailure bool
	flag.BoolVar(&ignoreFailure, "ignore-failure", false, "Exits with 0 status in the event of malformed input from user-data")

	var file string
	flag.StringVar(&file, "from-file", "", "Read user-data from provided file")

	var url string
	flag.StringVar(&url, "from-url", "", "Download user-data from provided url")

	var workspace string
	flag.StringVar(&workspace, "workspace", "/var/lib/coreos-cloudinit", "Base directory coreos-cloudinit should use to store data")

	var sshKeyName string
	flag.StringVar(&sshKeyName, "ssh-key-name", initialize.DefaultSSHKeyName, "Add SSH keys to the system with the given name")

	flag.Parse()

	if printVersion == true {
		fmt.Printf("coreos-cloudinit version %s\n", version)
		os.Exit(0)
	}

	if file != "" && url != "" {
		fmt.Println("Provide one of --from-file or --from-url")
		os.Exit(1)
	}

	var ds datasource.Datasource
	if file != "" {
		ds = datasource.NewLocalFile(file)
	} else if url != "" {
		ds = datasource.NewMetadataService(url)
	} else {
		fmt.Println("Provide one of --from-file or --from-url")
		os.Exit(1)
	}

	log.Printf("Fetching user-data from datasource of type %q", ds.Type())
	userdata, err := ds.Fetch()
	if err != nil {
		log.Printf("Failed fetching user-data from datasource: %v", err)
		if ignoreFailure {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	if len(userdata) == 0 {
		log.Printf("No user data to handle, exiting.")
		os.Exit(0)
	}

	parsed, err := ParseUserData(userdata)
	if err != nil {
		log.Printf("Failed parsing user-data: %v", err)
		if ignoreFailure {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	env := initialize.NewEnvironment("/", workspace)
	err = initialize.PrepWorkspace(env.Workspace())
	if err != nil {
		log.Fatalf("Failed preparing workspace: %v", err)
	}

	switch t := parsed.(type) {
	case initialize.CloudConfig:
		err = initialize.Apply(t, env)
	case system.Script:
		var path string
		path, err = initialize.PersistScriptInWorkspace(t, env.Workspace())
		if err == nil {
			var name string
			name, err = system.ExecuteScript(path)
			initialize.PersistUnitNameInWorkspace(name, workspace)
		}
	}

	if err != nil {
		log.Fatalf("Failed resolving user-data: %v", err)
	}
}

func ParseUserData(contents []byte) (interface{}, error) {
	bytereader := bytes.NewReader(contents)
	bufreader := bufio.NewReader(bytereader)
	header, _ := bufreader.ReadString('\n')

	if strings.HasPrefix(header, "#!") {
		log.Printf("Parsing user-data as script")
		return system.Script(contents), nil

	} else if header == "#cloud-config\n" {
		log.Printf("Parsing user-data as cloud-config")
		cfg, err := initialize.NewCloudConfig(contents)
		if err != nil {
			log.Fatal(err.Error())
		}
		return *cfg, nil
	} else {
		return nil, fmt.Errorf("Unrecognized user-data header: %s", header)
	}
}
