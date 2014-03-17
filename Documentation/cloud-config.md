# Customize CoreOS with Cloud-Config

CoreOS allows you to configure machine parameters, launch systemd units on startup and more. Only a subset of [cloud-config functionality][cloud-config] is implemented. A set of custom parameters were added to the cloud-config format that are specific to CoreOS.

[cloud-config]: http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data

## Supported cloud-config Parameters

### ssh_authorized_keys

Provided public SSH keys will be authorized for the `core` user.

The keys will be named "coreos-cloudinit" by default.
Override this with the `--ssh-key-name` flag when calling `coreos-cloudinit`.

### hostname

The provided value will be used to set the system's hostname.
This is the local part of a fully-qualified domain name (i.e. `foo` in `foo.example.com`).

### users

Add or modify users with the `users` directive by providing a list of user objects, each consisting of the following fields.
Each field is optional and of type string unless otherwise noted.
All but the `passwd` and `ssh-authorized-keys` fields will be ignored if the user already exists.

- **name**: Required. Login name of user
- **gecos**: GECOS comment of user
- **passwd**: Hash of the password to use for this user
- **homedir**: User's home directory. Defaults to /home/<name>
- **no-create-home**: Boolean. Skip home directory createion.
- **primary-group**: Default group for the user. Defaults to a new group created named after the user.
- **groups**: Add user to these additional groups
- **no-user-group**: Boolean. Skip default group creation.
- **ssh-authorized-keys**: List of public SSH keys to authorize for this user
- **system**: Create the user as a system user. No home directory will be created.
- **no-log-init**: Boolean. Skip initialization of lastlog and faillog databases.

The following fields are not yet implemented:

- **inactive**: Deactivate the user upon creation
- **lock-passwd**: Boolean. Disable password login for user
- **sudo**: Entry to add to /etc/sudoers for user. By default, no sudo access is authorized.
- **selinux-user**: Corresponding SELinux user
- **ssh-import-id**: Import SSH keys by ID from Launchpad.

##### Generating a password hash

Generating a safe hash is important to the security of your system.  Currently with updated tools like [oclhashcat](http://hashcat.net/oclhashcat/) simplified hashes like md5crypt are trivial to crack on modern GPU hardware.  You can generate a "safer" hash (read: not safe, never publish your hashes publicly) via:

###### On Debian/Ubuntu (via the package "whois")
    mkpasswd --method=SHA-512 --rounds=4096

###### With OpenSSL (note: this will only make md5crypt.  While better than plantext it should not be considered fully secure)
    openssl passwd -1

###### With Python (change password and salt values)
    python -c "import crypt, getpass, pwd; print crypt.crypt('password', '\$6\$SALT\$')"

###### With Perl (change password and salt values)
    perl -e 'print crypt("password","\$6\$SALT\$") . "\n"'

Using a higher number of rounds will help create more secure passwords, but given enough time, password hashes can be reversed.  On most RPM based distributions there is a tool called mkpasswd available in the `expect` package, but this does not handle "rounds" nor advanced hashing algorithms. 

### write_files

Inject an arbitrary set of files to the local filesystem.
Provide a list of objects with the following attributes:

- **path**: Absolute location on disk where contents should be written
- **content**: Data to write at the provided `path`
- **permissions**: String representing file permissions in octal notation (i.e. '0644')
- **owner**: User and group that should own the file written to disk. This is equivalent to the `<user>:<group>` argument to `chown <user>:<group> <path>`.

## Custom cloud-config Parameters

### coreos.oem

These fields are borrowed from the [os-release spec][os-release] and repurposed
as a way for cloud-init to know about the OEM partition on this machine.

- **id**: A lower case string identifying the oem.
- **version-id**: A lower case string identifying the version of the OEM. Example: `168.0.0`
- **name**: A name without the version that is suitable for presentation to the user.
- **home-url**: Link to the homepage of the provider or OEM.
- **bug-report-url***: Link to a place to file bug reports about this OEM partition.

cloudinit must render these fields down to an /etc/oem-release file on disk in the following format:

```
NAME=Rackspace
ID=rackspace
VERSION_ID=168.0.0
PRETTY_NAME="Rackspace Cloud Servers"
HOME_URL="http://www.rackspace.com/cloud/servers/"
BUG_REPORT_URL="https://github.com/coreos/coreos-overlay"
```

[os-release]: http://www.freedesktop.org/software/systemd/man/os-release.html

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

### Add a user

```
#cloud-config

users:
  - name: elroy
	passwd: $6$5s2u6/jR$un0AvWnqilcgaNB3Mkxd5yYv6mTlWfOoCYHZmfi3LDKVltj.E8XNKEcwWm...
	groups:
	  - staff
	  - docker
	ssh-authorized-keys:
	  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0g+ZTxC7weoIJLUafOgrm+h...
```

### Inject configuration files

```
#cloud-config

write_files:
  - path: /etc/hosts
    contents: |
      127.0.0.1    localhost
      192.0.2.211  buildbox
  - path: /etc/resolv.conf
    contents: |
      nameserver 192.0.2.13
      nameserver 192.0.2.14
```
