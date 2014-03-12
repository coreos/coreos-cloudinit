# coreos-cloudinit

coreos-cloudinit allows a user to customize CoreOS machines by providing either an executable script or a cloud-config document as instance user-data. See below to learn how to use these features.


## cloud-config

Only a subset of [cloud-config functionality][cloud-config] is implemented. A set of custom parameters were added to the cloud-config format that are specific to CoreOS.

[cloud-config]: http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data

### Supported cloud-config Parameters

#### ssh_authorized_keys

Provided public SSH keys will be authorized for the `core` user.

The keys will be named "coreos-cloudinit" by default.
Override this with the `--ssh-key-name` flag when calling `coreos-cloudinit`.

### Custom cloud-config Parameters

#### coreos.etcd.discovery_url

The value of `coreos.etcd.discovery_url` will be used to discover the instance's etcd peers using the [etcd discovery protocol][disco-proto]. Usage of the [public discovery service][disco-service] is encouraged.

[disco-proto]: https://github.com/coreos/etcd/blob/master/Documentation/discovery-protocol.md
[disco-service]: http://discovery.etcd.io

#### coreos.update

The `coreos.update` namespace has two keys:

- `coreos.update.server` is the auto update URL.
- `coreos.update.group` is the group your machine will join.

These fields will be written out to and replace `/etc/coreos/update.conf`. If only one of the parameters is given it will only overwrite the given field.

## user-data Script

Simply set your user-data to a script where the first line is a shebang:

```
#!/bin/bash

echo 'Hello, world!'
```

## Examples

### Inject an SSH key, bootstrap etcd, and start fleet using a cloud-config

```
#cloud-config

coreos:
	etcd:
		discovery_url: https://discovery.etcd.io/827c73219eeb2fa5530027c37bf18877
    fleet:
        autostart: yes

ssh_authorized_keys:
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0g+ZTxC7weoIJLUafOgrm+h...
```

