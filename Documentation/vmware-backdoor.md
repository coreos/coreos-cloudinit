# VMware Backdoor #

coreos-cloudinit is capable of reading userdata and metadata from the VMware
backdoor. This datasource can be enable with the `--from-vmware-backdoor` flag.
Userdata and metadata are passed from the hypervisor to the virtual machine
through guest variables. The following guest variables and their expected types
are supported by coreos-cloudinit:

|            guest variable             |              type               |
|:--------------------------------------|:--------------------------------|
| `hostname`                            | `hostname`                      |
| `interface.<n>.name`                  | `string`                        |
| `interface.<n>.mac`                   | `MAC address`                   |
| `interface.<n>.dhcp`                  | `{"yes", "no"}`                 |
| `interface.<n>.role`                  | `{"public", "private"}`         |
| `interface.<n>.ip.<m>.address`        | `CIDR IP address`               |
| `interface.<n>.route.<l>.gateway`     | `IP address`                    |
| `interface.<n>.route.<l>.destination` | `CIDR IP address`               |
| `dns.server.<x>`                      | `IP address`                    |
| `coreos.config.data`                  | `string`                        |
| `coreos.config.data.encoding`         | `{"", "base64", "gzip+base64"}` |

Note: "n", "m", "l", and "x" are 0-indexed, incrementing integers. The
identifier for the interfaces does not correspond to anything outside of this
configuration; it is merely for mapping configuration values to each interface.
