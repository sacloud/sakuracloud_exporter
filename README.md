# sakuracloud_exporter

![Test Status](https://github.com/sacloud/sakuracloud_exporter/workflows/Tests/badge.svg)
[![Slack](https://img.shields.io/badge/Slack-Sacloud%20Workspace-brightgreen)](https://join.slack.com/t/sacloud/shared_invite/zt-ohsdpv2t-_Bi1_F0jDAAmWjoMQCmAxg)  
[![License](https://img.shields.io/github/license/sacloud/sakuracloud_exporter)](LICENSE)
[![Version](https://img.shields.io/github/v/tag/sacloud/sakuracloud_exporter)](https://github.com/sacloud/sakuracloud_exporter/releases/latest)
![Downloads](https://img.shields.io/github/downloads/sacloud/sakuracloud_exporter/total)

[Prometheus](https://prometheus.io) exporter for [SakuraCloud](https://cloud.sakura.ad.jp) metrics.

## Installation

### Binaries

Download the already existing [binaries](https://github.com/sacloud/sakuracloud_exporter/releases/latest) for your platform:

```bash
$ ./sakuracloud_exporter <flags> 
```

If you want to use sakuracloud_exporter with systemd, see [systemd examples](examples/systemd).

### From source

Using the standard `go install` (you must have [Go][golang] already installed in your local machine):

```bash
$ go install github.com/sacloud/sakuracloud_exporter

$ sakuracloud_exporter <flags>
```

### Docker

To run the SakuraCloud exporter as a Docker container, run:

```bash
$ docker run -p 9542:9542 -e SAKURACLOUD_ACCESS_TOKEN=<YOUR-TOKEN> -e SAKURACLOUD_ACCESS_TOKEN_SECRET=<YOUR-SECRET> ghcr.io/sacloud/sakuracloud_exporter 
```

## Usage

### Flags

| Flag / Environment Variable                    | Required | Default    | Description                                                     |
| --------------------------------------------   | -------- | ---------- | -------------------------                                       |
| `--token` / `SAKURACLOUD_ACCESS_TOKEN`         | ◯       |            | API Key(Token)                                                  |
| `--secret` / `SAKURACLOUD_ACCESS_TOKEN_SECRET` | ◯       |            | API Key(Secret)                                                 |
| `--ratelimit`/ `SAKURACLOUD_RATE_LIMIT`        |          | `5`        | API request rate limit(maximum:10)                              |
| `--webaddr` / `WEB_ADDR`                       |          | `:9542`    | Exporter's listen address                                       |
| `--webpath`/ `WEB_PATH`                        |          | `/metrics` | Metrics request path                                            |
| `--no-collector.auto-backup`                   |          | `false`    | Disable the AutoBackup collector                                |
| `--no-collector.coupon`                        |          | `false`    | Disable the Coupon collector                                    |
| `--no-collector.database`                      |          | `false`    | Disable the Database collector                                  |
| `--no-collector.esme`                          |          | `false`    | Disable the ESME collector                                      |
| `--no-collector.internet`                      |          | `false`    | Disable the Internet(Switch+Router) collector                   |
| `--no-collector.load-balancer`                 |          | `false`    | Disable the LoadBalancer collector                              |
| `--no-collector.local-router`                  |          | `false`    | Disable the LocalRouter collector                               |
| `--no-collector.mobile-gateway`                |          | `false`    | Disable the MobileGateway collector                             |
| `--no-collector.nfs`                           |          | `false`    | Disable the NFS collector                                       |
| `--no-collector.proxy-lb`                      |          | `false`    | Disable the ProxyLB(Enhanced LoadBalancer) collector            |
| `--no-collector.server`                        |          | `false`    | Disable the Server collector                                    |
| `--no-collector.server.except-maintenance`     |          | `false`    | Disable the Server collector except for maintenance information |
| `--no-collector.sim`                           |          | `false`    | Disable the SIM collector                                       |
| `--no-collector.vpc-router`                    |          | `false`    | Disable the VPCRouter collector                                 |
| `--no-collector.zone`                          |          | `false`    | Disable the Zone collector                                      |


#### Flags for debug

| Flag / Environment Variable                  | Required | Default    | Description                 |
| -------------------------------------------- | -------- | ---------- | -------------------------   |
| `--fake-mode` / `FAKE_MODE`                     |          |            | The file path of fake store. If set this, make enabled to fake-store-mode(powered by libsacloud's fake driver) |

Example fake store file(JSON) is here[examples/fake/generate-fake-store-json/example-fake-store.json].

### Metrics

#### Supported Resource Types

The exporter returns the following metrics:

| Resource Type                   | Metric Name Prefix           |
| ------                          | -----------                  |
| [AutoBackup](#autobackup)       | sakuracloud_auto_backup_*    |
| [Coupon](#coupon)               | sakuracloud_coupon_*         |
| [Database](#database)           | sakuracloud_database_*       |
| [ESME](#esme)                   | sakuracloud_esme_*           |
| [Switch+Router](#switchrouter)  | sakuracloud_internet_*       |
| [LoadBalancer](#loadbalancer)   | sakuracloud_loadbalancer_*   |
| [LocalRouter](#localrouter)     | sakuracloud_local_router_*   |
| [MobileGateway](#mobilegateway) | sakuracloud_mobile_gateway_* |
| [NFS](#nfs)                     | sakuracloud_nfs_*            |
| [ProxyLB](#proxylb)             | sakuracloud_proxylb_*        |
| [Server](#server)               | sakuracloud_server_*         |
| [SIM](#sim)                     | sakuracloud_sim_*            |
| [VPCRouter](#vpcrouter)         | sakuracloud_vpc_router_*     |
| [Zone](#zone)                   | sakuracloud_zone_*           |
| [Exporter](#exporter)           | sakuracloud_exporter_*       |


#### AutoBackup

| Metric                               | Description                                                                | Labels                                                                                       |
| ------                               | -----------                                                                | ------                                                                                       |
| sakuracloud_auto_backup_info         | A metric with a constant '1' value labeled by auto_backup information      | `id`, `name`, `disk_id`, `max_backup_num`, `weekdays`, `tags`, `descriptions`                |
| sakuracloud_auto_backup_count        | A count of archives created by AutoBackup                                  | `id`, `name`, `disk_id`                                                                      |
| sakuracloud_auto_backup_last_time    | Last backup time in seconds since epoch (1970)                             | `id`, `name`, `disk_id`                                                                      |
| sakuracloud_auto_backup_archive_info | A metric with a constant '1' value labeled by backuped archive information | `id`, `name`, `disk_id`, `archive_id`, `archive_name`, `archive_tags`, `archive_description` |

#### Coupon

| Metric                            | Description                                          | Labels                           |
| ------                            | -----------                                          | ------                           |
| sakuracloud_coupon_discount       | The balance of coupon                                | `id`, `member_id`, `contract_id` |
| sakuracloud_coupon_remaining_days | The count of coupon's remaining days                 | `id`, `member_id`, `contract_id` |
| sakuracloud_coupon_exp_date       | Coupon expiration date in seconds since epoch (1970) | `id`, `member_id`, `contract_id` |
| sakuracloud_coupon_usable         | 1 if coupon is usable                                | `id`, `member_id`, `contract_id` |

#### Database

| Metric                                 | Description                                                        | Labels                                                                                                                                                                     |
| ------                                 | -----------                                                        | ------                                                                                                                                                                     |
| sakuracloud_database_info              | A metric with a constant '1' value labeled by database information | `id`, `name`, `zone`, `plan`, `host`, `database_type`, `database_revision`, `database_version`, `web_ui`, `replication_enabled`, `replication_role`, `tags`, `description` |
| sakuracloud_database_up                | If 1 the database is up and running, 0 otherwise                   | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_cpu_time          | Database's CPU time(unit:ms)                                       | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_memory_used       | Database's used memory size(unit:GB)                               | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_memory_total      | Database's total memory size(unit:GB)                              | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_nic_info          | A metric with a constant '1' value labeled by nic information      | `id`, `name`, `zone`, `upstream_type`, `upstream_id`, `upstream_name`, `ipaddress`, `nw_mask_len`, `gateway`                                                               |
| sakuracloud_database_nic_receive       | NIC's receive bytes(unit: Kbps)                                    | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_nic_send          | NIC's send bytes(unit: Kbps)                                       | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_disk_system_used  | Database's used system-disk size(unit:GB)                          | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_disk_system_total | Database's total system-disk size(unit:GB)                         | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_disk_backup_used  | Database's used backup-disk size(unit:GB)                          | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_disk_backup_total | Database's total backup-disk size(unit:GB)                         | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_binlog_used       | Database's used binlog size(unit:GB)                               | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_disk_read         | Disk's read bytes(unit: KBps)                                      | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_disk_write        | Disk's write bytes(unit: KBps)                                     | `id`, `name`, `zone`                                                                                                                                                       |
| sakuracloud_database_replication_delay | Replication delay time(unit:second)                                | `id`, `name`, `zone`                                                                                                                                                       |

#### ESME

| Metric                               | Description                                                     | Labels                               |
| ------                               | -----------                                                     | ------                               |
| sakuracloud_esme_info                | A metric with a constant '1' value labeled by ESME information  | `id`, `name`, `tags`, `description`  |
| sakuracloud_esme_message_count       | A count of messages handled by ESME                             | `id`, `name`, `status`               |


#### Switch+Router

| Metric                       | Description                                                        | Labels                                                                |
| ------                       | -----------                                                        | ------                                                                |
| sakuracloud_internet_info    | A metric with a constant '1' value labeled by internet information | `id`, `name`, `zone`, `switch_id`, `bandwidth`, `tags`, `description` |
| sakuracloud_internet_receive | Total receive bytes(unit: Kbps)                                    | `id`, `name`, `zone`, `switch_id`                                     |
| sakuracloud_internet_send    | Total send bytes(unit: Kbps)                                       | `id`, `name`, `zone`, `switch_id`                                     |

#### LoadBalancer

| Metric                                     | Description                                                            | Labels                                                                                                                  |
| ------                                     | -----------                                                            | ------                                                                                                                  |
| sakuracloud_loadbalancer_info              | A metric with a constant '1' value labeled by loadbalancer information | `id`, `name`, `zone`, `plan`, `ha`, `vrid`, `ipaddress1`, `ipaddress2`, `gateway`, `nw_mask_len`, `tags`, `description` |
| sakuracloud_loadbalancer_up                | If 1 the loadbalancer is up and running, 0 otherwise                   | `id`, `name`, `zone`                                                                                                    |
| sakuracloud_loadbalancer_receive           | Loadbalancer's receive bytes(unit: Kbps)                               | `id`, `name`, `zone`                                                                                                    |
| sakuracloud_loadbalancer_send              | Loadbalancer's receive bytes(unit: Kbps)                               | `id`, `name`, `zone`                                                                                                    |
| sakuracloud_loadbalancer_vip_info          | A metric with a constant '1' value labeld by vip information           | `id`, `name`, `zone`, `vip_index`, `vip`, `port`, `interval`, `sorry_server`, `description`                             |
| sakuracloud_loadbalancer_vip_cps           | Connection count per second                                            | `id`, `name`, `zone`, `vip_index`, `vip`                                                                                |
| sakuracloud_loadbalancer_server_info       | A metric with a constant '1' value labeld by real-server information   | `id`, `name`, `zone`, `vip_index`, `vip`, `server_index`, `ipaddress` ,`monitor`, `path`, `response_code`               |
| sakuracloud_loadbalancer_server_up         | If 1 the server is up and running, 0 otherwise                         | `id`, `name`, `zone`, `vip_index`, `vip`, `server_index`, `ipaddress`                                                   |
| sakuracloud_loadbalancer_server_connection | Current connection count                                               | `id`, `name`, `zone`, `vip_index`, `vip`, `server_index`, `ipaddress`                                                   |
| sakuracloud_loadbalancer_server_cps        | Connection count per second                                            | `id`, `name`, `zone`, `vip_index`, `vip`, `server_index`, `ipaddress`                                                   |

#### LocalRouter

| Metric                                     | Description                                                                            | Labels                                                                 |
| ------                                     | -----------                                                                            | ------                                                                 |
| sakuracloud_local_router_info              | A metric with a constant '1' value labeled by localRouter information                  | `id`, `name`, `tags`, `description`                                    |
| sakuracloud_local_router_up                | If 1 the localRouter is up and running, 0 otherwise                                    | `id`, `name`                                                           |
| sakuracloud_local_router_switch_info       | A metric with a constant '1' value labeled by localRouter connected switch information | `id`, `name`, `category`, `code`, `zone_id`                            |
| sakuracloud_local_router_network_info      | A metric with a constant '1' value labeled by network information of the localRouter   | `id`, `name`, `vip`, `ipaddress1`, `ipaddress2`, `nw_mask_len`, `vrid` |
| sakuracloud_local_router_static_route_info | A metric with a constant '1' value labeled by static route information                 | `id`, `name`, `route_index`, `prefix`, `next_hop`                      |
| sakuracloud_local_router_peer_info         | A metric with a constant '1' value labeled by peer information                         | `id`, `name`, `peer_index`, `peer_id`, `enabled`, `description`        |
| sakuracloud_local_router_peer_up           | If 1 the Peer is available, 0 otherwise                                                | `id`, `name`, `peer_index`, `peer_id`                                  |
| sakuracloud_local_router_receive_per_sec   | Receive bytes per seconds                                                              | `id`, `name`                                                           |
| sakuracloud_local_router_send_per_sec      | Send bytes per seconds                                                                 | `id`, `name`                                                           |

#### MobileGateway

| Metric                                          | Description                                                               | Labels                                                                                                                                       |
| ------                                          | -----------                                                               | ------                                                                                                                                       |
| sakuracloud_mobile_gateway_info                 | A metric with a constant '1' value labeled by mobile_gateway information  | `id`, `name`, `zone`, `internet_connection`, `inter_device_communication`, `tags`, `description`                                             |
| sakuracloud_mobile_gateway_up                   | If 1 the mobile_gateway is up and running, 0 otherwise                    | `id`, `name`, `zone`                                                                                                                         |
| sakuracloud_mobile_gateway_nic_receive          | MobileGateway's receive bytes(unit: Kbps)                                 | `id`, `name`, `zone`, `nic_index`, `ipaddress`, `nw_mask_len`                                                                                |
| sakuracloud_mobile_gateway_nic_send             | MobileGateway's send bytes(unit: Kbps)                                    | `id`, `name`, `zone`, `nic_index`, `ipaddress`, `nw_mask_len`                                                                                |
| sakuracloud_mobile_gateway_traffic_control_info | A metric with a constant '1' value labeled by traffic-control information | `id`, `name`, `zone` , `traffic_quota_in_mb`, `bandwidth_limit_in_kbps`, `enable_email`, `enable_slack`, `slack_url`, `auto_traffic_shaping` |
| sakuracloud_mobile_gateway_traffic_uplink       | MobileGateway's uplink bytes(unit: KB)                                    | `id`, `name`, `zone`                                                                                                                         |
| sakuracloud_mobile_gateway_traffic_downlink     | MobileGateway's downlink bytes(unit: KB)                                  | `id`, `name`, `zone`                                                                                                                         |
| sakuracloud_mobile_gateway_traffic_shaping      | If 1 the traffic is shaped, 0 otherwise                                   | `id`, `name`, `zone`                                                                                                                         |

#### NFS

| Metric                         | Description                                                   | Labels                                                                                      |
| ------                         | -----------                                                   | ------                                                                                      |
| sakuracloud_nfs_info           | A metric with a constant '1' value labeled by nfs information | `id`, `name`, `zone`, `plan`, `size`, `host`, `tags`, `description`                                 |
| sakuracloud_nfs_up             | If 1 the nfs is up and running, 0 otherwise                   | `id`, `name`, `zone`                                                                        |
| sakuracloud_nfs_free_disk_size | NFS's Free Disk Size(unit: GB)                                | `id`, `name`, `zone`                                                                        |
| sakuracloud_nfs_nic_info       | A metric with a constant '1' value labeled by nic information | `id`, `name`, `zone`, `upstream_id`, `upstream_name`, `ipaddress`, `nw_mask_len`, `gateway` |
| sakuracloud_nfs_receive        | NIC's receive bytes(unit: Kbps)                               | `id`, `name`, `zone`                                                                        |
| sakuracloud_nfs_send           | NIC's send bytes(unit: Kbps)                                  | `id`, `name`, `zone`                                                                        |

#### Server

| Metric                                   | Description                                                           | Labels                                                                                                                                                         |
| ------                                   | -----------                                                           | ------                                                                                                                                                         |
| sakuracloud_server_info                  | A metric with a constant '1' value labeled by server information      | `id`, `name`, `zone`, `cpus`, `disks`, `nics`, `memories`, `host`, `tags`, `description`                                                                       |
| sakuracloud_server_up                    | If 1 the server is up and running, 0 otherwise                        | `id`, `name`, `zone`                                                                                                                                           |
| sakuracloud_server_cpus                  | Number of server's vCPU cores                                         | `id`, `name`, `zone`                                                                                                                                           |
| sakuracloud_server_cpu_time              | Server's CPU time(unit: ms)                                           | `id`, `name`, `zone`                                                                                                                                           |
| sakuracloud_server_memories              | Size of server's memories(unit: GB)                                   | `id`, `name`, `zone`                                                                                                                                           |
| sakuracloud_server_disk_info             | A metric with a constant '1' value labeled by disk information        | `id`, `name`, `zone`, `disk_id`, `disk_name`, `index`, `plan`, `interface`, `size`, `tags`, `description`, `storage_id`, `storage_class`, `storage_generation` |
| sakuracloud_server_disk_read             | Disk's read bytes(unit: KBps)                                         | `id`, `name`, `zone`, `disk_id`, `disk_name`, `index`                                                                                                          |
| sakuracloud_server_disk_write            | Disk's write bytes(unit: KBps)                                        | `id`, `name`, `zone`, `disk_id`, `disk_name`, `index`                                                                                                          |
| sakuracloud_server_nic_info              | A metric with a constant '1' value labeled by nic information         | `id`, `name`, `zone`, `interface_id`, `index`, `upstream_type`, `upstream_id`, `upstream_name`                                                                 |
| sakuracloud_server_nic_bandwidth         | NIC's Bandwidth(unit: Mbps)                                           | `id`, `name`, `zone`, `interface_id`, `index`                                                                                                                  |
| sakuracloud_server_nic_receive           | NIC's receive bytes(unit: Kbps)                                       | `id`, `name`, `zone`, `interface_id`, `index`                                                                                                                  |
| sakuracloud_server_nic_send              | NIC's send bytes(unit: Kbps)                                          | `id`, `name`, `zone`, `interface_id`, `index`                                                                                                                  |
| sakuracloud_server_maintenance_info      | A metric with a constant '1' value labeled by maintenance information | `id`, `name`, `zone`, `info_url`, `info_title`, `description`, `start_date`, `end_date`                                                                        |
| sakuracloud_server_maintenance_scheduled | If 1 the server has scheduled maintenance info, 0 otherwise           | `id`, `name`, `zone`                                                                                                                                           |
| sakuracloud_server_maintenance_start     | Scheduled maintenance start time in seconds since epoch (1970)        | `id`, `name`, `zone`                                                                                                                                           |
| sakuracloud_server_maintenance_end       | Scheduled maintenance end time in seconds since epoch (1970)          | `id`, `name`, `zone`                                                                                                                                           |

#### ProxyLB

| Metric                                 | Description                                                           | Labels                                                                                                       |
| ------                                 | -----------                                                           | ------                                                                                                       |
| sakuracloud_proxylb_info               | A metric with a constant '1' value labeled by proxyLB information     | `plan`, `vip`, `fqdn`, `proxy_networks`, `sorry_server_ipaddress`, `sorry_server_port`, `tags`, `description` |
| sakuracloud_proxylb_up                 | If 1 the ProxyLB is available, 0 otherwise                            | `id`, `name`                                                                                                 |
| sakuracloud_proxylb_bind_port_info     | A metric with a constant '1' value labeled by BindPort information    | `id`, `name`, `bind_port_index`, `proxy_mode`, `port`                                                        |
| sakuracloud_proxylb_server_info        | A metric with a constant '1' value labeled by real-server information | `id`, `name`, `server_index`, `ipaddress`, `port`, `enabled`                                                 |
| sakuracloud_proxylb_cert_info          | A metric with a constant '1' value labeled by certificate information | `id`, `name`, `cert_index`, `common_name`, `issuer_name`                                                      |
| sakuracloud_proxylb_cert_expire        | Certificate expiration date in seconds since epoch (1970)             | `id`, `name`, `cert_index`                                                                                   |
| sakuracloud_proxylb_active_connections | Active connection count                                               | `id`, `name`                                                                                                 |
| sakuracloud_proxylb_connection_per_sec | Connection count per second                                           | `id`, `name`                                                                                                 |

#### SIM

| Metric                                | Description                                                   | Labels                                                                                                                                           |
| ------                                | -----------                                                   | ------                                                                                                                                           |
| sakuracloud_sim_info                  | A metric with a constant '1' value labeled by sim information | `id`, `name`, `imei_lock`, `registered_date`, `activated_date`, `deactivated_date`, `ipaddress`, `simgroup_id`, `carriers`, `tags`, `description` |
| sakuracloud_sim_session_up            | If 1 the session is up and running, 0 otherwise               | `id`, `name`                                                                                                                                     |
| sakuracloud_sim_uplink                | Uplink traffic (unit: Kbps)                                   | `id`, `name`                                                                                                                                     |
| sakuracloud_sim_downlink              | Downlink traffic (unit: Kbps)                                 | `id`, `name`                                                                                                                                     |

#### VPCRouter

| Metric                                  | Description                                                          | Labels                                                                                                                                     |
|-----------------------------------------|----------------------------------------------------------------------| ------                                                                                                                                     |
| sakuracloud_vpc_router_info             | A metric with a constant '1' value labeled by vpc_router information | `id`, `name`, `zone`, `plan`, `ha`, `vrid`, `vip`, `ipaddress1`, `ipaddress2`, `nw_mask_len`, `internet_connection`, `tags`, `description` |
| sakuracloud_vpc_router_up               | If 1 the vpc_router is up and running, 0 otherwise                   | `id`, `name`, `zone`                                                                                                                       |
| sakuracloud_vpc_router_cpu_time         | VPCRouter's CPU time(unit: ms)                                       | `id`, `name`, `zone`                                                                                                                       |
| sakuracloud_vpc_router_session          | Current session count                                                | `id`, `name`, `zone`                                                                                                                       |
| sakuracloud_vpc_router_dhcp_lease       | Current DHCPServer lease count                                       | `id`, `name`, `zone`                                                                                                                       |
| sakuracloud_vpc_router_l2tp_session     | Current L2TP-IPsec session count                                     | `id`, `name`, `zone`                                                                                                                       |
| sakuracloud_vpc_router_pptp_session     | Current PPTP session count                                           | `id`, `name`, `zone`                                                                                                                       |
| sakuracloud_vpc_router_s2s_peer_up      | If 1 the vpc_router's site to site peer is up, 0 otherwise           | `id`, `name`, `zone`, `peer_address`, `peer_index`                                                                                         |
| sakuracloud_vpc_router_session_analysis | Session statistics for VPC routers                                   | `id`, `name`, `zone`, `type`, `label`                                                                                                      |
| sakuracloud_vpc_router_receive          | VPCRouter's receive bytes(unit: Kbps)                                | `id`, `name`, `zone`, `nic_index`, `vip`, `ipaddress1`, `ipaddress2`, `nw_mask_len`                                                        |
| sakuracloud_vpc_router_send             | VPCRouter's receive bytes(unit: Kbps)                                | `id`, `name`, `zone`, `nic_index`, `vip`, `ipaddress1`, `ipaddress2`, `nw_mask_len`                                                        |

#### Zone

| Metric                | Description                                                    | Labels                                                  |
| ------                | -----------                                                    | ------                                                  |
| sakuracloud_zone_info | A metric with a constant '1' value labeled by zone information | `id`, `name`, `description`, `region_id`, `region_name` |

#### Exporter

| Metric                            | Description                                                                | Labels                             |
| ------                            | -----------                                                                | ------                             |
| sakuracloud_exporter_start_time   | Unix timestamp of the start time                                           | -                                  |
| sakuracloud_exporter_build_info   | A metric with a constant '1' value labeled by exporter's build information | `version`, `revision`, `goversion` |
| sakuracloud_exporter_errors_total | The total number of errors per collector                                   | `collector`                        |

## License

 `sakuracloud_exporter` Copyright (C) 2019-2022 sakuracloud_exporter authors.

  This project is published under [Apache 2.0 License](LICENSE).
  
