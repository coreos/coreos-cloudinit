package cloudinit

import (
	"fmt"
	"log"

	"github.com/coreos/coreos-cloudinit/third_party/launchpad.net/goyaml"
)

const DefaultSSHKeyName = "coreos-cloudinit"

type CloudConfig struct {
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	Coreos            struct {
		Etcd  struct{ Discovery_URL string }
		Fleet struct{ Autostart bool }
		Units []Unit
	}
	WriteFiles []WriteFile `yaml:"write_files"`
	Hostname   string
	Users      []User
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

func ApplyCloudConfig(cfg CloudConfig, sshKeyName string) error {
	if cfg.Hostname != "" {
		if err := SetHostname(cfg.Hostname); err != nil {
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

			if UserExists(&user) {
				log.Printf("User '%s' exists, ignoring creation-time fields", user.Name)
				if user.PasswordHash != "" {
					log.Printf("Setting '%s' user's password", user.Name)
					if err := SetUserPassword(user.Name, user.PasswordHash); err != nil {
						log.Printf("Failed setting '%s' user's password: %v", user.Name, err)
						return err
					}
				}
			} else {
				log.Printf("Creating user '%s'", user.Name)
				if err := CreateUser(&user); err != nil {
					log.Printf("Failed creating user '%s': %v", user.Name, err)
					return err
				}
			}

			if len(user.SSHAuthorizedKeys) > 0 {
				log.Printf("Authorizing %d SSH keys for user '%s'", len(user.SSHAuthorizedKeys), user.Name)
				if err := AuthorizeSSHKeys(user.Name, sshKeyName, user.SSHAuthorizedKeys); err != nil {
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
		err := AuthorizeSSHKeys("core", sshKeyName, cfg.SSHAuthorizedKeys)
		if err == nil {
			log.Printf("Authorized SSH keys for core user")
		} else {
			return err
		}
	}

	if len(cfg.WriteFiles) > 0 {
		for _, file := range cfg.WriteFiles {
			if err := ProcessWriteFile("/", &file); err != nil {
				return err
			}
			log.Printf("Wrote file %s to filesystem", file.Path)
		}
	}

	if cfg.Coreos.Etcd.Discovery_URL != "" {
		err := PersistEtcdDiscoveryURL(cfg.Coreos.Etcd.Discovery_URL)
		if err == nil {
			log.Printf("Consumed etcd discovery url")
		} else {
			log.Fatalf("Failed to persist etcd discovery url to filesystem: %v", err)
		}
	}

	if len(cfg.Coreos.Units) > 0 {
		for _, unit := range cfg.Coreos.Units {
			log.Printf("Placing unit %s on filesystem", unit.Name)
			dst, err := PlaceUnit("/", &unit)
			if err != nil {
				return err
			}
			log.Printf("Placed unit %s at %s", unit.Name, dst)

			if unit.Group() != "network" {
				log.Printf("Enabling unit file %s", dst)
				if err := EnableUnitFile(dst, unit.Runtime); err != nil {
					return err
				}
				log.Printf("Enabled unit %s", unit.Name)
			} else {
				log.Printf("Skipping enable for network-like unit %s", unit.Name)
			}
		}
		DaemonReload()
		StartUnits(cfg.Coreos.Units)
	}

	if cfg.Coreos.Fleet.Autostart {
		err := StartUnitByName("fleet.service")
		if err == nil {
			log.Printf("Started fleet service.")
		} else {
			return err
		}
	}

	return nil
}
