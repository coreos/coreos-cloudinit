# Customize with Cloud-Config

CoreOS allows you to configure networking, create users, launch systemd units on startup and more. We've designed our implementation to allow the same cloud-config file to work across all of our supported platforms.

Only a subset of [cloud-config functionality][cloud-config] is implemented. A set of custom parameters were added to the cloud-config format that are specific to CoreOS. An example file containing all available options can be found at the bottom of this page.

[cloud-config]: http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data

## CoreOS Parameters

### coreos.etcd

The `coreos.etcd.*` options are translated to a partial systemd unit acting as an etcd configuration file.
We can use the templating feature of coreos-cloudinit to automate etcd configuration with the `$private_ipv4` and `$public_ipv4` fields. For example, the following cloud-config document...

```
#cloud-config

coreos:
    etcd:
        name: node001
        discovery: https://discovery.etcd.io/3445fa65423d8b04df07f59fb40218f8
        addr: $public_ipv4:4001
        peer-addr: $private_ipv4:7001
```

...will generate a systemd unit drop-in like this:

```
[Service]
Environment="ETCD_NAME=node001"
Environment="ETCD_DISCOVERY=https://discovery.etcd.io/3445fa65423d8b04df07f59fb40218f8"
Environment="ETCD_ADDR=203.0.113.29:4001"
Environment="ETCD_PEER_ADDR=192.0.2.13:7001"
```

For more information about the available configuration options, see the [etcd documentation][etcd-config].
Note that hyphens in the coreos.etcd.* keys are mapped to underscores.

[etcd-config]: https://github.com/coreos/etcd/blob/master/Documentation/configuration.md

### coreos.oem

These fields are borrowed from the [os-release spec][os-release] and repurposed
as a way for coreos-cloudinit to know about the OEM partition on this machine:

- **id**: Lowercase string identifying the OEM
- **name**: Human-friendly string representing the OEM
- **version-id**: Lowercase string identifying the version of the OEM
- **home-url**: Link to the homepage of the provider or OEM
- **bug-report-url**: Link to a place to file bug reports about this OEM

coreos-cloudinit renders these fields to `/etc/oem-release`.
If no **id** field is provided, coreos-cloudinit will ignore this section.

For example, the following cloud-config document...

```
#cloud-config
coreos:
  oem:
    id: rackspace
    name: Rackspace Cloud Servers
    version-id: 168.0.0
    home-url: https://www.rackspace.com/cloud/servers/
    bug-report-url: https://github.com/coreos/coreos-overlay
```

...would be rendered to the following `/etc/oem-release`:

```
ID="rackspace"
NAME="Rackspace Cloud Servers"
VERSION_ID="168.0.0"
HOME_URL="https://www.rackspace.com/cloud/servers/"
BUG_REPORT_URL="https://github.com/coreos/coreos-overlay"
```

[os-release]: http://www.freedesktop.org/software/systemd/man/os-release.html

### coreos.units

Arbitrary systemd units may be provided in the `coreos.units` attribute.
`coreos.units` is a list of objects with the following fields:

- **name**: String representing unit's name. Required.
- **runtime**: Boolean indicating whether or not to persist the unit across reboots. This is analagous to the `--runtime` argument to `systemd enable`. Default value is false.
- **content**: Plaintext string representing entire unit file. If no value is provided, the unit is assumed to exist already.
- **command**: Command to execute on unit: start, stop, reload, restart, try-restart, reload-or-restart, reload-or-try-restart. Default value is restart.

**NOTE:** The command field is ignored for all network, netdev, and link units. The systemd-networkd.service unit will be restarted in their place.

##### Examples

Write a unit to disk, automatically starting it.

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

Start the builtin `etcd` and `fleet` services:

```
# cloud-config

coreos:
    units:
      - name: etcd.service
        command: start
      - name: fleet.service
        command: start
```

## Cloud-Config Parameters

### ssh_authorized_keys

Provided public SSH keys will be authorized for the `core` user.

The keys will be named "coreos-cloudinit" by default.
Override this with the `--ssh-key-name` flag when calling `coreos-cloudinit`.

```
#cloud-config

ssh_authorized_keys:
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0g+ZTxC7weoIJLUafOgrm+h...
```

### hostname

The provided value will be used to set the system's hostname.
This is the local part of a fully-qualified domain name (i.e. `foo` in `foo.example.com`).

```
#cloud-config

hostname: coreos1
```

### users

Add or modify users with the `users` directive by providing a list of user objects, each consisting of the following fields.
Each field is optional and of type string unless otherwise noted.
All but the `passwd` and `ssh-authorized-keys` fields will be ignored if the user already exists.

- **name**: Required. Login name of user
- **gecos**: GECOS comment of user
- **passwd**: Hash of the password to use for this user
- **homedir**: User's home directory. Defaults to /home/<name>
- **no-create-home**: Boolean. Skip home directory creation.
- **primary-group**: Default group for the user. Defaults to a new group created named after the user.
- **groups**: Add user to these additional groups
- **no-user-group**: Boolean. Skip default group creation.
- **ssh-authorized-keys**: List of public SSH keys to authorize for this user
- **coreos-ssh-import-github**: Authorize SSH keys from Github user
- **coreos-ssh-import-url**: Authorize SSH keys imported from a url endpoint.
- **system**: Create the user as a system user. No home directory will be created.
- **no-log-init**: Boolean. Skip initialization of lastlog and faillog databases.

The following fields are not yet implemented:

- **inactive**: Deactivate the user upon creation
- **lock-passwd**: Boolean. Disable password login for user
- **sudo**: Entry to add to /etc/sudoers for user. By default, no sudo access is authorized.
- **selinux-user**: Corresponding SELinux user
- **ssh-import-id**: Import SSH keys by ID from Launchpad.

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

#### Generating a password hash

If you choose to use a password instead of an SSH key, generating a safe hash is extremely important to the security of your system. Simplified hashes like md5crypt are trivial to crack on modern GPU hardware. Here are a few ways to generate secure hashes:

```
# On Debian/Ubuntu (via the package "whois")
mkpasswd --method=SHA-512 --rounds=4096

# OpenSSL (note: this will only make md5crypt.  While better than plantext it should not be considered fully secure)
openssl passwd -1

# Python (change password and salt values)
python -c "import crypt, getpass, pwd; print crypt.crypt('password', '\$6\$SALT\$')"

# Perl (change password and salt values)
perl -e 'print crypt("password","\$6\$SALT\$") . "\n"'
```

Using a higher number of rounds will help create more secure passwords, but given enough time, password hashes can be reversed.  On most RPM based distributions there is a tool called mkpasswd available in the `expect` package, but this does not handle "rounds" nor advanced hashing algorithms. 

#### Retrieving ssh authorized keys from a GitHub user

Using the field `coreos-ssh-import-github` you can make coreos-cloudinit to add the public ssh keys from a GitHub user as authorized keys to a server.

```
#cloud-config

users:
  - name: elroy
    coreos-ssh-import-github: elroy
```

#### Retrieving ssh authorized keys from an http endpoint

coreos-cloudinit can also pull public SSH keys from any http endpoint that matches [GitHub's API response format](https://developer.github.com/v3/users/keys/#list-public-keys-for-a-user).
For example, if you have an installation of GitHub Enterprise, you can provide a complete url with an authentication token:

```
#cloud-config

users:
  - name: elroy
    coreos-ssh-import-url: https://token:<OAUTH-TOKEN>@github-enterprise.example.com/users/elroy/keys
```

You can also provide any url which response matches that json format for public keys:

```
#cloud-config

users:
  - name: elroy
    coreos-ssh-import-url: https://example.com/public-keys
```

### write_files

Inject an arbitrary set of files to the local filesystem.
Provide a list of objects with the following attributes:

- **path**: Absolute location on disk where contents should be written
- **content**: Data to write at the provided `path`
- **permissions**: String representing file permissions in octal notation (i.e. '0644')
- **owner**: User and group that should own the file written to disk. This is equivalent to the `<user>:<group>` argument to `chown <user>:<group> <path>`.
