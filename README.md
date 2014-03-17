# coreos-cloudinit

coreos-cloudinit enables a user to customize CoreOS machines by providing either an executable script or a cloud-config document as instance user-data.

## Supported Cloud-Config Features

A subset of [cloud-config][cloud-config] is implemented in coreos-cloudinit and is [documented here](https://github.com/coreos/coreos-cloudinit/tree/master/Documentation/cloud-config.md). In addition specific CoreOS paramaters were added for unit files, etcd discovery urls, and others.

[cloud-config]: http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data
