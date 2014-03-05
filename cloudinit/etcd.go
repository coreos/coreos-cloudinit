package cloudinit

import (
	"io/ioutil"
	"log"
	"os"
	"path"
)

const (
	etcdDiscoveryPath = "/var/run/etcd/bootstrap.disco"
)

func PersistEtcdDiscoveryURL(url string) error {
	dir := path.Dir(etcdDiscoveryPath)
	if _, err := os.Stat(dir); err != nil {
		log.Printf("Creating directory /var/run/etcd")
		err := os.MkdirAll(dir, os.FileMode(0644))
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(etcdDiscoveryPath, []byte(url), os.FileMode(0644))
}
