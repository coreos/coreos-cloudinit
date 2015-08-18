# Cloud-Config Locations

On every boot, coreos-cloudinit looks for a config file to configure your host. Here is a list of locations which are used by the Cloud-Config utility, depending on your CoreOS platform:

* Mount point with [config-2](/os/docs/latest/config-drive.html#contents-and-format) label. It should contain a `openstack/latest/user_data` relative path. An absolute path inside CoreOS should look like `/media/configvirtfs/openstack/latest/user_data`. Usually used by cloud providers or in VM installations.
* FAT or ISO9660 filesystem with [config-2](/os/docs/latest/config-drive.html#qemu-virtfs) label. A `/media/configdrive/` mount point should contain a `/media/configdrive/openstack/latest/user_data` absolute path to the config file. Usually used in VM installations.
* Kernel command line: `cloud-config-url=http://example.com/user_data`. You can find this string using this command `cat /proc/cmdline`. Usually used in [PXE](/os/docs/latest/booting-with-pxe.html) or [iPXE](/os/docs/latest/booting-with-ipxe.html) boots.
* `/var/lib/coreos-install/user_data` when you install CoreOS manually using the [coreos-install](/os/docs/latest/installing-to-disk.html) tool. Usually used in bare metal installations.
* OEM images use `/usr/share/oem/cloud-config.yml` path.
* Azure platform uses OEM path for first Cloud-Config initialization and then `/var/lib/waagent/CustomData` to apply your settings.
* DigitalOcean and EC2 use URLs to download Cloud-Config.

You can also run the `coreos-cloudinit` tool manually and provide a path to your custom Cloud-Config file:

```sh
sudo coreos-cloudinit --from-file=/home/core/cloud-config.yaml
``` 

This command will apply your custom cloud-config.
