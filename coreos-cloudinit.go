package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/coreos/coreos-cloudinit/cloudinit"
)

func main() {
	var userdata []byte
	var err error

	var file string
	flag.StringVar(&file, "from-file", "", "Read user-data from file rather than metadata service")

	var workspace string
	flag.StringVar(&workspace, "workspace", "/var/lib/coreos-cloudinit", "Base directory coreos-cloudinit should use to store data")

	flag.Parse()

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
