# Systemd Unit

The unit file in this directory is to be put into `/etc/systemd/system`.
It needs a user named `sakuracloud_exporter`, whose shell should be `/sbin/nologin` and should not have any special privileges.
(e.g. `useradd -s /sbin/nologin -M sakuracloud_exporter`)

It needs a sysconfig file in `/etc/sysconfig/sakuracloud_exporter`.
A sample file can be found in `sysconfig.sakuracloud_exporter`.
