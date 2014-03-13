package cloudinit

import (
	"fmt"
	"log"

	"github.com/coreos/coreos-cloudinit/third_party/launchpad.net/goyaml"
)

const DefaultSSHKeyName = "coreos-cloudinit"

type CloudConfig struct {
	SSH_Authorized_Keys []string
	Coreos              struct {
		Etcd  struct{ Discovery_URL string }
		Fleet struct{ Autostart bool }
		Units []Unit
	}
	Write_Files []WriteFile
	Hostname    string
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

	if len(cfg.SSH_Authorized_Keys) > 0 {
		err := AuthorizeSSHKeys(sshKeyName, cfg.SSH_Authorized_Keys)
		if err == nil {
			log.Printf("Authorized SSH keys for core user")
		} else {
			return err
		}
	}

	if len(cfg.Write_Files) > 0 {
		for _, file := range cfg.Write_Files {
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
