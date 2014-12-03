/*
   Copyright 2014 CoreOS, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package initialize

import (
	"errors"
	"fmt"
	"log"
	"path"

	"github.com/coreos/coreos-cloudinit/Godeps/_workspace/src/gopkg.in/yaml.v1"

	"github.com/coreos/coreos-cloudinit/network"
	"github.com/coreos/coreos-cloudinit/system"
)

// CloudConfigFile represents a CoreOS specific configuration option that can generate
// an associated system.File to be written to disk
type CloudConfigFile interface {
	// File should either return (*system.File, error), or (nil, nil) if nothing
	// needs to be done for this configuration option.
	File(root string) (*system.File, error)
}

// CloudConfigUnit represents a CoreOS specific configuration option that can generate
// associated system.Units to be created/enabled appropriately
type CloudConfigUnit interface {
	Units(root string) ([]system.Unit, error)
}

// CloudConfig encapsulates the entire cloud-config configuration file and maps directly to YAML
type CloudConfig struct {
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	Coreos            struct {
		Etcd   EtcdEnvironment
		Fleet  FleetEnvironment
		OEM    OEMRelease
		Update UpdateConfig
		Units  []system.Unit
	}
	WriteFiles        []system.File `yaml:"write_files"`
	Hostname          string
	Users             []system.User
	ManageEtcHosts    EtcHosts `yaml:"manage_etc_hosts"`
	NetworkConfigPath string
	NetworkConfig     string
}

type warner func(format string, v ...interface{})

// warnOnUnrecognizedKeys parses the contents of a cloud-config file and calls
// warn(msg, key) for every unrecognized key (i.e. those not present in CloudConfig)
func warnOnUnrecognizedKeys(contents string, warn warner) {
	// Generate a map of all understood cloud config options
	var cc map[string]interface{}
	b, _ := yaml.Marshal(&CloudConfig{})
	yaml.Unmarshal(b, &cc)

	// Now unmarshal the entire provided contents
	var c map[string]interface{}
	yaml.Unmarshal([]byte(contents), &c)

	// Check that every key in the contents exists in the cloud config
	for k, _ := range c {
		if _, ok := cc[k]; !ok {
			warn("Warning: unrecognized key %q in provided cloud config - ignoring section", k)
		}
	}

	// Check for unrecognized coreos options, if any are set
	if coreos, ok := c["coreos"]; ok {
		if set, ok := coreos.(map[interface{}]interface{}); ok {
			known := cc["coreos"].(map[interface{}]interface{})
			for k, _ := range set {
				if key, ok := k.(string); ok {
					if _, ok := known[key]; !ok {
						warn("Warning: unrecognized key %q in coreos section of provided cloud config - ignoring", key)
					}
				} else {
					warn("Warning: unrecognized key %q in coreos section of provided cloud config - ignoring", k)
				}
			}
		}
	}

	// Check for any badly-specified users, if any are set
	if users, ok := c["users"]; ok {
		var known map[string]interface{}
		b, _ := yaml.Marshal(&system.User{})
		yaml.Unmarshal(b, &known)

		if set, ok := users.([]interface{}); ok {
			for _, u := range set {
				if user, ok := u.(map[interface{}]interface{}); ok {
					for k, _ := range user {
						if key, ok := k.(string); ok {
							if _, ok := known[key]; !ok {
								warn("Warning: unrecognized key %q in user section of cloud config - ignoring", key)
							}
						} else {
							warn("Warning: unrecognized key %q in user section of cloud config - ignoring", k)
						}
					}
				}
			}
		}
	}

	// Check for any badly-specified files, if any are set
	if files, ok := c["write_files"]; ok {
		var known map[string]interface{}
		b, _ := yaml.Marshal(&system.File{})
		yaml.Unmarshal(b, &known)

		if set, ok := files.([]interface{}); ok {
			for _, f := range set {
				if file, ok := f.(map[interface{}]interface{}); ok {
					for k, _ := range file {
						if key, ok := k.(string); ok {
							if _, ok := known[key]; !ok {
								warn("Warning: unrecognized key %q in file section of cloud config - ignoring", key)
							}
						} else {
							warn("Warning: unrecognized key %q in file section of cloud config - ignoring", k)
						}
					}
				}
			}
		}
	}
}

// NewCloudConfig instantiates a new CloudConfig from the given contents (a
// string of YAML), returning any error encountered. It will ignore unknown
// fields but log encountering them.
func NewCloudConfig(contents string) (*CloudConfig, error) {
	var cfg CloudConfig
	err := yaml.Unmarshal([]byte(contents), &cfg)
	if err != nil {
		return &cfg, err
	}
	warnOnUnrecognizedKeys(contents, log.Printf)
	return &cfg, nil
}

func (cc CloudConfig) String() string {
	bytes, err := yaml.Marshal(cc)
	if err != nil {
		return ""
	}

	stringified := string(bytes)
	stringified = fmt.Sprintf("#cloud-config\n%s", stringified)

	return stringified
}

// Apply renders a CloudConfig to an Environment. This can involve things like
// configuring the hostname, adding new users, writing various configuration
// files to disk, and manipulating systemd services.
func Apply(cfg CloudConfig, env *Environment) error {
	if cfg.Hostname != "" {
		if err := system.SetHostname(cfg.Hostname); err != nil {
			return err
		}
		log.Printf("Set hostname to %s", cfg.Hostname)
	}

	for _, user := range cfg.Users {
		if user.Name == "" {
			log.Printf("User object has no 'name' field, skipping")
			continue
		}

		if system.UserExists(&user) {
			log.Printf("User '%s' exists, ignoring creation-time fields", user.Name)
			if user.PasswordHash != "" {
				log.Printf("Setting '%s' user's password", user.Name)
				if err := system.SetUserPassword(user.Name, user.PasswordHash); err != nil {
					log.Printf("Failed setting '%s' user's password: %v", user.Name, err)
					return err
				}
			}
		} else {
			log.Printf("Creating user '%s'", user.Name)
			if err := system.CreateUser(&user); err != nil {
				log.Printf("Failed creating user '%s': %v", user.Name, err)
				return err
			}
		}

		if len(user.SSHAuthorizedKeys) > 0 {
			log.Printf("Authorizing %d SSH keys for user '%s'", len(user.SSHAuthorizedKeys), user.Name)
			if err := system.AuthorizeSSHKeys(user.Name, env.SSHKeyName(), user.SSHAuthorizedKeys); err != nil {
				return err
			}
		}
		if user.SSHImportGithubUser != "" {
			log.Printf("Authorizing github user %s SSH keys for CoreOS user '%s'", user.SSHImportGithubUser, user.Name)
			if err := SSHImportGithubUser(user.Name, user.SSHImportGithubUser); err != nil {
				return err
			}
		}
		if user.SSHImportURL != "" {
			log.Printf("Authorizing SSH keys for CoreOS user '%s' from '%s'", user.Name, user.SSHImportURL)
			if err := SSHImportKeysFromURL(user.Name, user.SSHImportURL); err != nil {
				return err
			}
		}
	}

	if len(cfg.SSHAuthorizedKeys) > 0 {
		err := system.AuthorizeSSHKeys("core", env.SSHKeyName(), cfg.SSHAuthorizedKeys)
		if err == nil {
			log.Printf("Authorized SSH keys for core user")
		} else {
			return err
		}
	}

	for _, ccf := range []CloudConfigFile{cfg.Coreos.OEM, cfg.Coreos.Update, cfg.ManageEtcHosts} {
		f, err := ccf.File(env.Root())
		if err != nil {
			return err
		}
		if f != nil {
			cfg.WriteFiles = append(cfg.WriteFiles, *f)
		}
	}

	for _, ccu := range []CloudConfigUnit{cfg.Coreos.Etcd, cfg.Coreos.Fleet, cfg.Coreos.Update} {
		u, err := ccu.Units(env.Root())
		if err != nil {
			return err
		}
		cfg.Coreos.Units = append(cfg.Coreos.Units, u...)
	}

	wroteEnvironment := false
	for _, file := range cfg.WriteFiles {
		fullPath, err := system.WriteFile(&file, env.Root())
		if err != nil {
			return err
		}
		if path.Clean(file.Path) == "/etc/environment" {
			wroteEnvironment = true
		}
		log.Printf("Wrote file %s to filesystem", fullPath)
	}

	if !wroteEnvironment {
		ef := env.DefaultEnvironmentFile()
		if ef != nil {
			err := system.WriteEnvFile(ef, env.Root())
			if err != nil {
				return err
			}
			log.Printf("Updated /etc/environment")
		}
	}

	if env.NetconfType() != "" {
		var interfaces []network.InterfaceGenerator
		var err error
		switch env.NetconfType() {
		case "debian":
			interfaces, err = network.ProcessDebianNetconf(cfg.NetworkConfig)
		case "digitalocean":
			interfaces, err = network.ProcessDigitalOceanNetconf(cfg.NetworkConfig)
		default:
			return fmt.Errorf("Unsupported network config format %q", env.NetconfType())
		}

		if err != nil {
			return err
		}

		if err := system.WriteNetworkdConfigs(interfaces); err != nil {
			return err
		}
		if err := system.RestartNetwork(interfaces); err != nil {
			return err
		}
	}

	um := system.NewUnitManager(env.Root())
	return processUnits(cfg.Coreos.Units, env.Root(), um)

}

// processUnits takes a set of Units and applies them to the given root using
// the given UnitManager. This can involve things like writing unit files to
// disk, masking/unmasking units, or invoking systemd
// commands against units. It returns any error encountered.
func processUnits(units []system.Unit, root string, um system.UnitManager) error {
	type action struct {
		unit    string
		command string
	}
	actions := make([]action, 0, len(units))
	reload := false
	restartNetworkd := false
	for _, unit := range units {
		dst := unit.Destination(root)
		if unit.Content != "" {
			log.Printf("Writing unit %s to filesystem at path %s", unit.Name, dst)
			if err := um.PlaceUnit(&unit, dst); err != nil {
				return err
			}
			log.Printf("Placed unit %s at %s", unit.Name, dst)
			reload = true
		}

		if unit.Mask {
			log.Printf("Masking unit file %s", unit.Name)
			if err := um.MaskUnit(&unit); err != nil {
				return err
			}
		} else if unit.Runtime {
			log.Printf("Ensuring runtime unit file %s is unmasked", unit.Name)
			if err := um.UnmaskUnit(&unit); err != nil {
				return err
			}
		}

		if unit.Enable {
			if unit.Group() != "network" {
				log.Printf("Enabling unit file %s", unit.Name)
				if err := um.EnableUnitFile(unit.Name, unit.Runtime); err != nil {
					return err
				}
				log.Printf("Enabled unit %s", unit.Name)
			} else {
				log.Printf("Skipping enable for network-like unit %s", unit.Name)
			}
		}

		if unit.Group() == "network" {
			restartNetworkd = true
		} else if unit.Command != "" {
			actions = append(actions, action{unit.Name, unit.Command})
		}
	}

	if reload {
		if err := um.DaemonReload(); err != nil {
			return errors.New(fmt.Sprintf("failed systemd daemon-reload: %v", err))
		}
	}

	if restartNetworkd {
		log.Printf("Restarting systemd-networkd")
		res, err := um.RunUnitCommand("restart", "systemd-networkd.service")
		if err != nil {
			return err
		}
		log.Printf("Restarted systemd-networkd (%s)", res)
	}

	for _, action := range actions {
		log.Printf("Calling unit command '%s %s'", action.command, action.unit)
		res, err := um.RunUnitCommand(action.command, action.unit)
		if err != nil {
			return err
		}
		log.Printf("Result of '%s %s': %s", action.command, action.unit, res)
	}

	return nil
}
