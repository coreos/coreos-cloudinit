# Customize CoreOS with Cloud-Config

CoreOS allows you to configure machine parameters, launch systemd units on startup and more. Only a subset of [cloud-config functionality][cloud-config] is implemented. A set of custom parameters were added to the cloud-config format that are specific to CoreOS.

[cloud-config]: http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data

## Supported cloud-config Parameters

### ssh_authorized_keys

Provided public SSH keys will be authorized for the `core` user.

The keys will be named "coreos-cloudinit" by default.
Override this with the `--ssh-key-name` flag when calling `coreos-cloudinit`.

## Custom cloud-config Parameters

### coreos.etcd.discovery_url

The value of `coreos.etcd.discovery_url` will be used to discover the instance's etcd peers using the [etcd discovery protocol][disco-proto]. Usage of the [public discovery service][disco-service] is encouraged.

[disco-proto]: https://github.com/coreos/etcd/blob/master/Documentation/discovery-protocol.md
[disco-service]: http://discovery.etcd.io

### coreos.units

Arbitrary systemd units may be provided in the `coreos.units` attribute.
`coreos.units` is a list of objects with the following fields:

- **name**: string representing unit's name
- **runtime**: boolean indicating whether or not to persist the unit across reboots. This is analagous to the `--runtime` flag to `systemd enable`.
- **content**: plaintext string representing entire unit file

See docker example below.

## user-data Script

Simply set your user-data to a script where the first line is a shebang:

```
#!/bin/bash

echo 'Hello, world!'
```

## Examples

### Inject an SSH key, bootstrap etcd, and start fleet
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

### Start a docker container on boot

```
#cloud-config

coreos:
    units:
      - name: docker-redis.service
        content: |
          [Unit]
          Description=Redis container
          Author=Me
          After=docker.service

          [Service]
          Restart=always
          ExecStart=/usr/bin/docker start -a redis_server
          ExecStop=/usr/bin/docker stop -t 2 redis_server
          
          [Install]
          WantedBy=local.target
```
