# coreos-cloudinit

coreos-cloudinit allows a user to customize CoreOS machines by providing either an executable script or a cloud-config document as instance user-data. See below to learn how to use these features.

## Supported Cloud-Config Features

Only a subset of [cloud-config functionality][cloud-config] is implemented. A set of custom parameters were added to the cloud-config format that are specific to CoreOS, which are [documented here](https://github.com/coreos/coreos-cloudinit/tree/master/Documentation/cloud-config.md).

[cloud-config]: http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data
