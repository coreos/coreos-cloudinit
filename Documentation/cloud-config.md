# Using Cloud-Config

CoreOS allows you to declaratively customize various OS-level items, such as network configuration, user accounts, and systemd units. This document describes the full list of items we can configure. The `coreos-cloudinit` program uses these files as it configures the OS after startup or during runtime. Your cloud-config is processed during each boot.

## Configuration File

The file used by this system initialization program is called a "cloud-config" file. It is inspired by the [cloud-init][cloud-init] project's [cloud-config][cloud-config] file. which is "the defacto multi-distribution package that handles early initialization of a cloud instance" ([cloud-init docs][cloud-init-docs]). Because the cloud-init project includes tools which aren't used by CoreOS, only the relevant subset of its configuration items will be implemented in our cloud-config file. In addition to those, we added a few CoreOS-specific items, such as etcd configuration, OEM definition, and systemd units.

We've designed our implementation to allow the same cloud-config file to work across all of our supported platforms.

[cloud-init]: https://launchpad.net/cloud-init
[cloud-init-docs]: http://cloudinit.readthedocs.org/en/latest/index.html
[cloud-config]: http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data

### File Format

The cloud-config file uses the [YAML][yaml] file format, which uses whitespace and new-lines to delimit lists, associative arrays, and values.

A cloud-config file should contain `#cloud-config`, followed by an associative array which has zero or more of the following keys:

- `coreos`
- `ssh_authorized_keys`
- `hostname`
- `users`
- `write_files`
- `manage_etc_hosts`

The expected values for these keys are defined in the rest of this document.

[yaml]: https://en.wikipedia.org/wiki/YAML

### Providing Cloud-Config with Config-Drive

CoreOS tries to conform to each platform's native method to provide user data. Each cloud provider tends to be unique, but this complexity has been abstracted by CoreOS. You can view each platform's instructions on their documentation pages. The most universal way to provide cloud-config is [via config-drive](https://github.com/coreos/coreos-cloudinit/blob/master/Documentation/config-drive.md), which attaches a read-only device to the machine, that contains your cloud-config file.

## Configuration Parameters

### coreos

#### etcd

The `coreos.etcd.*` parameters will be translated to a partial systemd unit acting as an etcd configuration file.
If the platform environment supports the templating feature of coreos-cloudinit it is possible to automate etcd configuration with the `$private_ipv4` and `$public_ipv4` fields. For example, the following cloud-config document...

```yaml
#cloud-config

coreos:
    etcd:
        name: node001
	# generate a new token for each unique cluster from https://discovery.etcd.io/new
        discovery: https://discovery.etcd.io/<token>
	# multi-region and multi-cloud deployments need to use $public_ipv4
        addr: $public_ipv4:4001
        peer-addr: $private_ipv4:7001
```

...will generate a systemd unit drop-in like this:

```yaml
[Service]
Environment="ETCD_NAME=node001"
Environment="ETCD_DISCOVERY=https://discovery.etcd.io/<token>"
Environment="ETCD_ADDR=203.0.113.29:4001"
Environment="ETCD_PEER_ADDR=192.0.2.13:7001"
```

For more information about the available configuration parameters, see the [etcd documentation][etcd-config].
Note that hyphens in the coreos.etcd.* keys are mapped to underscores.

_Note: The `$private_ipv4` and `$public_ipv4` substitution variables referenced in other documents are only supported on Amazon EC2, Google Compute Engine, OpenStack, Rackspace, and Vagrant._

[etcd-config]: https://github.com/coreos/etcd/blob/master/Documentation/configuration.md

#### fleet

The `coreos.fleet.*` parameters work very similarly to `coreos.etcd.*`, and allow for the configuration of fleet through environment variables. For example, the following cloud-config document...

```yaml
#cloud-config

coreos:
    fleet:
        public-ip: $public_ipv4
        metadata: region=us-west
```

...will generate a systemd unit drop-in like this:

```yaml
[Service]
Environment="FLEET_PUBLIC_IP=203.0.113.29"
Environment="FLEET_METADATA=region=us-west"
```

For more information on fleet configuration, see the [fleet documentation][fleet-config].

[fleet-config]: https://github.com/coreos/fleet/blob/master/Documentation/configuration.md

#### update

The `coreos.update.*` parameters manipulate settings related to how CoreOS instances are updated.

These fields will be written out to and replace `/etc/coreos/update.conf`. If only one of the parameters is given it will only overwrite the given field. 
The `reboot-strategy` parameter also affects the behaviour of [locksmith](https://github.com/coreos/locksmith). 

- **reboot-strategy**: One of "reboot", "etcd-lock", "best-effort" or "off" for controlling when reboots are issued after an update is performed.
  - _reboot_: Reboot immediately after an update is applied.
  - _etcd-lock_: Reboot after first taking a distributed lock in etcd, this guarantees that only one host will reboot concurrently and that the cluster will remain available during the update.
  - _best-effort_ - If etcd is running, "etcd-lock", otherwise simply "reboot".
  - _off_ - Disable rebooting after updates are applied (not recommended).
- **server**: is the omaha endpoint URL which will be queried for updates.
- **group**:  signifies the channel which should be used for automatic updates.  This value defaults to the version of the image initially downloaded. (one of "master", "alpha", "beta", "stable")

*Note: cloudinit will only manipulate the locksmith unit file in the systemd runtime directory (`/run/systemd/system/locksmithd.service`). If any manual modifications are made to an overriding unit configuration file (e.g. `/etc/systemd/system/locksmithd.service`), cloudinit will no longer be able to control the locksmith service unit.*

##### Example

```yaml
#cloud-config
coreos:
  update:
    reboot-strategy: etcd-lock
```

#### units

The `coreos.units.*` parameters define a list of arbitrary systemd units to start after booting. This feature is intended to help you start essential services required to mount storage and configure networking in order to join the CoreOS cluster. It is not intended to be a Chef/Puppet replacement.

Each item is an object with the following fields:

- **name**: String representing unit's name. Required.
- **runtime**: Boolean indicating whether or not to persist the unit across reboots. This is analogous to the `--runtime` argument to `systemctl enable`. Default value is false.
- **enable**: Boolean indicating whether or not to handle the [Install] section of the unit file. This is similar to running `systemctl enable <name>`. Default value is false.
- **content**: Plaintext string representing entire unit file. If no value is provided, the unit is assumed to exist already.
- **command**: Command to execute on unit: start, stop, reload, restart, try-restart, reload-or-restart, reload-or-try-restart. Default value is restart.
- **mask**: Whether to mask the unit file by symlinking it to `/dev/null` (analogous to `systemctl mask <name>`). Note that unlike `systemctl mask`, **this will destructively remove any existing unit file** located at `/etc/systemd/system/<unit>`, to ensure that the mask succeeds. Default value is false.

**NOTE:** The command field is ignored for all network, netdev, and link units. The systemd-networkd.service unit will be restarted in their place.

##### Examples

Write a unit to disk, automatically starting it.

```yaml
#cloud-config

coreos:
    units:
      - name: docker-redis.service
        command: start
        content: |
          [Unit]
          Description=Redis container
          Author=Me
          After=docker.service

          [Service]
          Restart=always
          ExecStart=/usr/bin/docker start -a redis_server
          ExecStop=/usr/bin/docker stop -t 2 redis_server
```

Start the built-in `etcd` and `fleet` services:

```yaml
#cloud-config

coreos:
    units:
      - name: etcd.service
        command: start
      - name: fleet.service
        command: start
```

### ssh_authorized_keys

The `ssh_authorized_keys` parameter adds public SSH keys which will be authorized for the `core` user.

The keys will be named "coreos-cloudinit" by default.
Override this by using the `--ssh-key-name` flag when calling `coreos-cloudinit`.

```yaml
#cloud-config

ssh_authorized_keys:
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0g+ZTxC7weoIJLUafOgrm+h...
```

### hostname

The `hostname` parameter defines the system's hostname.
This is the local part of a fully-qualified domain name (i.e. `foo` in `foo.example.com`).

```yaml
#cloud-config

hostname: coreos1
```

### users

The `users` parameter adds or modifies the specified list of users. Each user is an object which consists of the following fields. Each field is optional and of type string unless otherwise noted.
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

```yaml
#cloud-config

users:
  - name: elroy
    passwd: $6$5s2u6/jR$un0AvWnqilcgaNB3Mkxd5yYv6mTlWfOoCYHZmfi3LDKVltj.E8XNKEcwWm...
    groups:
      - sudo
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

#### Retrieving SSH Authorized Keys

##### From a GitHub User

Using the `coreos-ssh-import-github` field, we can import public SSH keys from a GitHub user to use as authorized keys to a server.

```yaml
#cloud-config

users:
  - name: elroy
    coreos-ssh-import-github: elroy
```

##### From an HTTP Endpoint

We can also pull public SSH keys from any HTTP endpoint which matches [GitHub's API response format](https://developer.github.com/v3/users/keys/#list-public-keys-for-a-user).
For example, if you have an installation of GitHub Enterprise, you can provide a complete URL with an authentication token:

```yaml
#cloud-config

users:
  - name: elroy
    coreos-ssh-import-url: https://github-enterprise.example.com/api/v3/users/elroy/keys?access_token=<TOKEN>
```

You can also specify any URL whose response matches the JSON format for public keys:

```yaml
#cloud-config

users:
  - name: elroy
    coreos-ssh-import-url: https://example.com/public-keys
```

### write_files

The `write-file` parameter defines a list of files to create on the local filesystem. Each file is represented as an associative array which has the following keys:

- **path**: Absolute location on disk where contents should be written
- **content**: Data to write at the provided `path`
- **permissions**: String representing file permissions in octal notation (i.e. '0644')
- **owner**: User and group that should own the file written to disk. This is equivalent to the `<user>:<group>` argument to `chown <user>:<group> <path>`.

Explicitly not implemented is the **encoding** attribute.
The **content** field must represent exactly what should be written to disk.

```yaml
#cloud-config
write_files:
  - path: /etc/fleet/fleet.conf
    permissions: 0644
    content: |
      verbosity=1
      metadata="region=us-west,type=ssd"
```

### manage_etc_hosts

The `manage_etc_hosts` parameter configures the contents of the `/etc/hosts` file, which is used for local name resolution.
Currently, the only supported value is "localhost" which will cause your system's hostname
to resolve to "127.0.0.1".  This is helpful when the host does not have DNS
infrastructure in place to resolve its own hostname, for example, when using Vagrant.

```yaml
#cloud-config

manage_etc_hosts: localhost
```
