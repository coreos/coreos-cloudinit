package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/initialize"
	"github.com/coreos/coreos-cloudinit/system"
)

const version = "0.7.1+git"

func init() {
	//Removes timestamp since it is displayed already during booting
	log.SetFlags(0)
}

func main() {
	var printVersion bool
	flag.BoolVar(&printVersion, "version", false, "Print the version and exit")

	var ignoreFailure bool
	flag.BoolVar(&ignoreFailure, "ignore-failure", false, "Exits with 0 status in the event of malformed input from user-data")

	var file string
	flag.StringVar(&file, "from-file", "", "Read user-data from provided file")

	var configdrive string
	flag.StringVar(&configdrive, "from-configdrive", "", "Read user-data from provided cloud-drive directory")

	var url string
	flag.StringVar(&url, "from-url", "", "Download user-data from provided url")

	var useProcCmdline bool
	flag.BoolVar(&useProcCmdline, "from-proc-cmdline", false, fmt.Sprintf("Parse %s for '%s=<url>', using the cloud-config served by an HTTP GET to <url>", datasource.ProcCmdlineLocation, datasource.ProcCmdlineCloudConfigFlag))

	var workspace string
	flag.StringVar(&workspace, "workspace", "/var/lib/coreos-cloudinit", "Base directory coreos-cloudinit should use to store data")

	var sshKeyName string
	flag.StringVar(&sshKeyName, "ssh-key-name", initialize.DefaultSSHKeyName, "Add SSH keys to the system with the given name")

	flag.Parse()

	if printVersion == true {
		fmt.Printf("coreos-cloudinit version %s\n", version)
		os.Exit(0)
	}

	var ds datasource.Datasource
	if file != "" {
		ds = datasource.NewLocalFile(file)
	} else if url != "" {
		ds = datasource.NewMetadataService(url)
	} else if configdrive != "" {
		ds = datasource.NewConfigDrive(configdrive)
	} else if useProcCmdline {
		ds = datasource.NewProcCmdline()
	} else {
		fmt.Println("Provide one of --from-file, --from-configdrive, --from-url or --from-proc-cmdline")
		os.Exit(1)
	}

	log.Printf("Fetching user-data from datasource of type %q", ds.Type())
	userdataBytes, err := ds.Fetch()
	if err != nil {
		log.Printf("Failed fetching user-data from datasource: %v", err)
		if ignoreFailure {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	if len(userdataBytes) == 0 {
		log.Printf("No user data to handle, exiting.")
		os.Exit(0)
	}

	env := initialize.NewEnvironment("/", workspace)

	userdata := string(userdataBytes)
	userdata = env.Apply(userdata)

	parsed, err := initialize.ParseUserData(userdata)
	if err != nil {
		log.Printf("Failed parsing user-data: %v", err)
		if ignoreFailure {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

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
