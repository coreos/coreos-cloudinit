package initialize

import (
	"fmt"
	"log"
	"path"

	"github.com/coreos/coreos-cloudinit/third_party/launchpad.net/goyaml"

	"github.com/coreos/coreos-cloudinit/system"
)

type CloudConfig struct {
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	Coreos            struct {
		Etcd  EtcdEnvironment
		Units []system.Unit
	}
	WriteFiles []system.File `yaml:"write_files"`
	Hostname   string
	Users      []system.User
}

func NewCloudConfig(contents []byte) (*CloudConfig, error) {
	var cfg CloudConfig
	err := goyaml.Unmarshal(contents, &cfg)
	return &cfg, err
}

func (cc CloudConfig) String() string {
	bytes, err := goyaml.Marshal(cc)
	if err != nil {
		return ""
	}

	stringified := string(bytes)
	stringified = fmt.Sprintf("#cloud-config\n%s", stringified)

	return stringified
}

func Apply(cfg CloudConfig, env *Environment) error {
	if cfg.Hostname != "" {
		if err := system.SetHostname(cfg.Hostname); err != nil {
			return err
		}
		log.Printf("Set hostname to %s", cfg.Hostname)
	}

	if len(cfg.Users) > 0 {
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

	if len(cfg.WriteFiles) > 0 {
		for _, file := range cfg.WriteFiles {
			file.Path = path.Join(env.Root(), file.Path)
			if err := system.WriteFile(&file); err != nil {
				return err
			}
			log.Printf("Wrote file %s to filesystem", file.Path)
		}
	}

	if len(cfg.Coreos.Etcd) > 0 {
		if err := WriteEtcdEnvironment(cfg.Coreos.Etcd, env.Root()); err != nil {
			log.Fatalf("Failed to write etcd config to filesystem: %v", err)
		}

		log.Printf("Wrote etcd config file to filesystem")
	}

	if len(cfg.Coreos.Units) > 0 {
		commands := make(map[string]string, 0)

		for _, unit := range cfg.Coreos.Units {
			if unit.Content != "" {
				log.Printf("Writing unit %s to filesystem", unit.Name)
				dst, err := system.PlaceUnit(&unit, env.Root())
				if err != nil {
					return err
				}
				log.Printf("Placed unit %s at %s", unit.Name, dst)

				if unit.Group() != "network" {
					log.Printf("Enabling unit file %s", dst)
					if err := system.EnableUnitFile(dst, unit.Runtime); err != nil {
						return err
					}
					log.Printf("Enabled unit %s", unit.Name)
				} else {
					log.Printf("Skipping enable for network-like unit %s", unit.Name)
				}
			}

			if unit.Group() != "network" {
				command := unit.Command
				if command == "" {
					command = "restart"
				}
				commands[unit.Name] = command
			} else {
				commands["systemd-networkd.service"] = "restart"
			}
		}

		system.DaemonReload()

		for unit, command := range commands {
			log.Printf("Calling unit command '%s %s'", command, unit)
			res, err := system.RunUnitCommand(command, unit)
			if err != nil {
				return err
			}
			log.Printf("Result of '%s %s': %s", command, unit, res)
		}
	}

	return nil
}
