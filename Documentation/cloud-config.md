# Customize CoreOS with Cloud-Config

CoreOS allows you to configure machine parameters, launch systemd units on startup and more. Only a subset of [cloud-config functionality][cloud-config] is implemented. A set of custom parameters were added to the cloud-config format that are specific to CoreOS.

[cloud-config]: http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data

## Supported cloud-config Parameters

### ssh_authorized_keys

Provided public SSH keys will be authorized for the `core` user.

The keys will be named "coreos-cloudinit" by default.
Override this with the `--ssh-key-name` flag when calling `coreos-cloudinit`.

#### users

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
