package main

import (
	"fmt"
	"flag"
	"io/ioutil"
	"os"
	"log"

	"github.com/coreos/coreos-cloudinit/cloudinit"
)

const version = "0.1.0+git"

func main() {
	var userdata []byte
	var err error

	var printVersion bool
	flag.BoolVar(&printVersion, "version", false, "Print the version and exit")

	var file string
	flag.StringVar(&file, "from-file", "", "Read user-data from file rather than metadata service")

	var workspace string
	flag.StringVar(&workspace, "workspace", "/var/lib/coreos-cloudinit", "Base directory coreos-cloudinit should use to store data")

	flag.Parse()

	if printVersion == true {
		fmt.Printf("coreos-cloudinit version %s\n", version)
		os.Exit(0)
	}

	if file != "" {
		log.Printf("Reading user-data from file: %s", file)
		userdata, err = ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		log.Printf("Reading user-data from metadata service")
		svc := cloudinit.NewMetadataService()
		userdata, err = svc.UserData()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	parsed, err := cloudinit.ParseUserData(userdata)
	if err != nil {
		log.Fatalf("Failed parsing user-data: %v", err)
	}

	err = cloudinit.PrepWorkspace(workspace)
	if err != nil {
		log.Fatalf("Failed preparing workspace: %v", err)
	}

	switch t := parsed.(type) {
	case cloudinit.CloudConfig:
		err = cloudinit.ResolveCloudConfig(t)
	case cloudinit.Script:
		var path string
		path, err = cloudinit.PersistScriptInWorkspace(t, workspace)
		if err == nil {
			err = cloudinit.ExecuteScript(path)
		}
	}

	if err != nil {
		log.Fatalf("Failed resolving user-data: %v", err)
	}
}
