# coreos-cloudinit

coreos-cloudinit enables a user to customize CoreOS machines by providing either a cloud-config document or an executable script through user-data.

## Configuration with cloud-config

A subset of the [official cloud-config spec][official-cloud-config] is implemented by coreos-cloudinit.
Additionally, several [CoreOS-specific options][custom-cloud-config] have been implemented to support interacting with unit files, bootstrapping etcd clusters, and more.
All supported cloud-config parameters are [documented here][all-cloud-config]. 

[official-cloud-config]: http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data
[custom-cloud-config]: https://github.com/coreos/coreos-cloudinit/blob/master/Documentation/cloud-config.md#coreos-parameters
[all-cloud-config]: https://github.com/coreos/coreos-cloudinit/tree/master/Documentation/cloud-config.md

The following is an example cloud-config document:

```
#cloud-config

coreos:
    units:
      - name: etcd.service
        command: start

users:
  - name: core
    passwd: $1$allJZawX$00S5T756I5PGdQga5qhqv1

write_files:
  - path: /etc/resolv.conf
    content: |
        nameserver 192.0.2.2
        nameserver 192.0.2.3
```


## Executing a Script

coreos-cloudinit supports executing user-data as a script instead of parsing it as a cloud-config document.
Make sure the first line of your user-data is a shebang and coreos-cloudinit will attempt to execute it:

```
#!/bin/bash

echo 'Hello, world!'
```
