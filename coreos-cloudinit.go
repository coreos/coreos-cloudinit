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
	version               = "0.9.1"
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

	// Extract IPv4 addresses from metadata if possible
	var subs map[string]string
	if len(metadataBytes) > 0 {
		subs, err = initialize.ExtractIPsFromMetadata(metadataBytes)
		if err != nil {
			fmt.Printf("Failed extracting IPs from meta-data: %v\n", err)
			die()
		}
	}

	// Apply environment to user-data
	env := initialize.NewEnvironment("/", ds.ConfigRoot(), workspace, convertNetconf, sshKeyName, subs)
	userdata := env.Apply(string(userdataBytes))

	var ccm, ccu *initialize.CloudConfig
	var script *system.Script
	if ccm, err = initialize.ParseMetaData(string(metadataBytes)); err != nil {
		fmt.Printf("Failed to parse meta-data: %v\n", err)
		die()
	}
	if ud, err := initialize.ParseUserData(userdata); err != nil {
		fmt.Printf("Failed to parse user-data: %v\n", err)
		die()
	} else {
		switch t := ud.(type) {
		case *initialize.CloudConfig:
			ccu = t
		case system.Script:
			script = &t
		}
	}

	var cc *initialize.CloudConfig
	if ccm != nil && ccu != nil {
		fmt.Println("Merging cloud-config from meta-data and user-data")
		merged := mergeCloudConfig(*ccm, *ccu)
		cc = &merged
	} else if ccm != nil && ccu == nil {
		fmt.Println("Processing cloud-config from meta-data")
		cc = ccm
	} else if ccm == nil && ccu != nil {
		fmt.Println("Processing cloud-config from user-data")
		cc = ccu
	} else {
		fmt.Println("No cloud-config data to handle.")
	}

	if cc != nil {
		if err = initialize.Apply(*cc, env); err != nil {
			fmt.Printf("Failed to apply cloud-config: %v\n", err)
			die()
		}
	}

	if script != nil {
		if err = runScript(*script, env); err != nil {
			fmt.Printf("Failed to run script: %v\n", err)
			die()
		}
	}
}

// mergeCloudConfig merges certain options from mdcc (a CloudConfig derived from
// meta-data) onto udcc (a CloudConfig derived from user-data), if they are
// not already set on udcc (i.e. user-data always takes precedence)
// NB: This needs to be kept in sync with ParseMetadata so that it tracks all
// elements of a CloudConfig which that function can populate.
func mergeCloudConfig(mdcc, udcc initialize.CloudConfig) (cc initialize.CloudConfig) {
	if mdcc.Hostname != "" {
		if udcc.Hostname != "" {
			fmt.Printf("Warning: user-data hostname (%s) overrides metadata hostname (%s)", udcc.Hostname, mdcc.Hostname)
		} else {
			udcc.Hostname = mdcc.Hostname
		}

	}
	for _, key := range mdcc.SSHAuthorizedKeys {
		udcc.SSHAuthorizedKeys = append(udcc.SSHAuthorizedKeys, key)
	}
	if mdcc.NetworkConfigPath != "" {
		if udcc.NetworkConfigPath != "" {
			fmt.Printf("Warning: user-data NetworkConfigPath %s overrides metadata NetworkConfigPath %s", udcc.NetworkConfigPath, mdcc.NetworkConfigPath)
		} else {
			udcc.NetworkConfigPath = mdcc.NetworkConfigPath
		}
	}
	return udcc
}

// getDatasources creates a slice of possible Datasources for cloudinit based
// on the different source command-line flags.
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

// selectDatasource attempts to choose a valid Datasource to use based on its
// current availability. The first Datasource to report to be available is
// returned. Datasources will be retried if possible if they are not
// immediately available. If all Datasources are permanently unavailable or
// datasourceTimeout is reached before one becomes available, nil is returned.
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

// TODO(jonboulle): this should probably be refactored and moved into a different module
func runScript(script system.Script, env *initialize.Environment) error {
	err := initialize.PrepWorkspace(env.Workspace())
	if err != nil {
		fmt.Printf("Failed preparing workspace: %v\n", err)
		return err
	}
	path, err := initialize.PersistScriptInWorkspace(script, env.Workspace())
	if err == nil {
		var name string
		name, err = system.ExecuteScript(path)
		initialize.PersistUnitNameInWorkspace(name, env.Workspace())
	}
	return err
}
