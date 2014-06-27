package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/initialize"
	"github.com/coreos/coreos-cloudinit/pkg"
	"github.com/coreos/coreos-cloudinit/system"
)

const (
	version               = "0.7.7+git"
	datasourceInterval    = 100 * time.Millisecond
	datasourceMaxInterval = 30 * time.Second
	datasourceTimeout     = 5 * time.Minute
)

var (
	printVersion  bool
	ignoreFailure bool
	sources       struct {
		file            string
		configDrive     string
		metadataService bool
		url             string
		procCmdLine     bool
	}
	convertNetconf string
	workspace      string
	sshKeyName     string
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "Print the version and exit")
	flag.BoolVar(&ignoreFailure, "ignore-failure", false, "Exits with 0 status in the event of malformed input from user-data")
	flag.StringVar(&sources.file, "from-file", "", "Read user-data from provided file")
	flag.StringVar(&sources.configDrive, "from-configdrive", "", "Read data from provided cloud-drive directory")
	flag.BoolVar(&sources.metadataService, "from-metadata-service", false, "Download data from metadata service")
	flag.StringVar(&sources.url, "from-url", "", "Download user-data from provided url")
	flag.BoolVar(&sources.procCmdLine, "from-proc-cmdline", false, fmt.Sprintf("Parse %s for '%s=<url>', using the cloud-config served by an HTTP GET to <url>", datasource.ProcCmdlineLocation, datasource.ProcCmdlineCloudConfigFlag))
	flag.StringVar(&convertNetconf, "convert-netconf", "", "Read the network config provided in cloud-drive and translate it from the specified format into networkd unit files (requires the -from-configdrive flag)")
	flag.StringVar(&workspace, "workspace", "/var/lib/coreos-cloudinit", "Base directory coreos-cloudinit should use to store data")
	flag.StringVar(&sshKeyName, "ssh-key-name", initialize.DefaultSSHKeyName, "Add SSH keys to the system with the given name")
}

func main() {
	flag.Parse()

	die := func() {
		if ignoreFailure {
			os.Exit(0)
		}
		os.Exit(1)
	}

	if printVersion == true {
		fmt.Printf("coreos-cloudinit version %s\n", version)
		os.Exit(0)
	}

	if convertNetconf != "" && sources.configDrive == "" {
		fmt.Println("-convert-netconf flag requires -from-configdrive")
		os.Exit(1)
	}

	switch convertNetconf {
	case "":
	case "debian":
	default:
		fmt.Printf("Invalid option to -convert-netconf: '%s'. Supported options: 'debian'\n", convertNetconf)
		os.Exit(1)
	}

	dss := getDatasources()
	if len(dss) == 0 {
		fmt.Println("Provide at least one of --from-file, --from-configdrive, --from-metadata-service, --from-url or --from-proc-cmdline")
		os.Exit(1)
	}

	ds := selectDatasource(dss)
	if ds == nil {
		fmt.Println("No datasources available in time")
		die()
	}

	fmt.Printf("Fetching user-data from datasource of type %q\n", ds.Type())
	userdataBytes, err := ds.FetchUserdata()
	if err != nil {
		fmt.Printf("Failed fetching user-data from datasource: %v\n", err)
		die()
	}

	fmt.Printf("Fetching meta-data from datasource of type %q\n", ds.Type())
	metadataBytes, err := ds.FetchMetadata()
	if err != nil {
		fmt.Printf("Failed fetching meta-data from datasource: %v\n", err)
		die()
	}

	var subs map[string]string
	if len(metadataBytes) > 0 {
		subs, err = initialize.ExtractIPsFromMetadata(metadataBytes)
		if err != nil {
			fmt.Printf("Failed extracting IPs from meta-data: %v\n", err)
			die()
		}
	}
	env := initialize.NewEnvironment("/", ds.ConfigRoot(), workspace, convertNetconf, sshKeyName, subs)

	if len(metadataBytes) > 0 {
		if err := processMetadata(string(metadataBytes), env); err != nil {
			fmt.Printf("Failed to process meta-data: %v\n", err)
			die()
		}
	} else {
		fmt.Println("No meta-data to handle.")
	}

	if len(userdataBytes) > 0 {
		if err := processUserdata(string(userdataBytes), env); err != nil {
			fmt.Printf("Failed to process user-data: %v\n", err)
			if !ignoreFailure {
				die()
			}
		}
	} else {
		fmt.Println("No user-data to handle.")
	}
}

func getDatasources() []datasource.Datasource {
	dss := make([]datasource.Datasource, 0, 5)
	if sources.file != "" {
		dss = append(dss, datasource.NewLocalFile(sources.file))
	}
	if sources.url != "" {
		dss = append(dss, datasource.NewRemoteFile(sources.url))
	}
	if sources.configDrive != "" {
		dss = append(dss, datasource.NewConfigDrive(sources.configDrive))
	}
	if sources.metadataService {
		dss = append(dss, datasource.NewMetadataService())
	}
	if sources.procCmdLine {
		dss = append(dss, datasource.NewProcCmdline())
	}
	return dss
}

func selectDatasource(sources []datasource.Datasource) datasource.Datasource {
	ds := make(chan datasource.Datasource)
	stop := make(chan struct{})
	var wg sync.WaitGroup

	for _, s := range sources {
		wg.Add(1)
		go func(s datasource.Datasource) {
			defer wg.Done()

			duration := datasourceInterval
			for {
				fmt.Printf("Checking availability of %q\n", s.Type())
				if s.IsAvailable() {
					ds <- s
					return
				} else if !s.AvailabilityChanges() {
					return
				}
				select {
				case <-stop:
					return
				case <-time.Tick(duration):
					duration = pkg.ExpBackoff(duration, datasourceMaxInterval)
				}
			}
		}(s)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	var s datasource.Datasource
	select {
	case s = <-ds:
	case <-done:
	case <-time.Tick(datasourceTimeout):
	}

	close(stop)
	return s
}

func processUserdata(userdata string, env *initialize.Environment) error {
	userdata = env.Apply(userdata)

	parsed, err := initialize.ParseUserData(userdata)
	if err != nil {
		fmt.Printf("Failed parsing user-data: %v\n", err)
		return err
	}

	err = initialize.PrepWorkspace(env.Workspace())
	if err != nil {
		fmt.Printf("Failed preparing workspace: %v\n", err)
		return err
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
			initialize.PersistUnitNameInWorkspace(name, env.Workspace())
		}
	}

	return err
}

func processMetadata(metadata string, env *initialize.Environment) error {
	parsed, err := initialize.ParseMetaData(metadata)
	if err != nil {
		fmt.Printf("Failed parsing meta-data: %v\n", err)
		return err
	}
	err = initialize.PrepWorkspace(env.Workspace())
	if err != nil {
		fmt.Printf("Failed preparing workspace: %v\n", err)
		return err
	}

	return initialize.Apply(parsed, env)
}
