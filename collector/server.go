// Copyright 2019-2020 The sakuracloud_exporter Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/sacloud/libsacloud/v2/utils/newsfeed"
	"github.com/sacloud/sakuracloud_exporter/iaas"
)

// ServerCollector collects metrics about all servers.
type ServerCollector struct {
	ctx    context.Context
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.ServerClient

	Up         *prometheus.Desc
	ServerInfo *prometheus.Desc
	CPUs       *prometheus.Desc
	CPUTime    *prometheus.Desc
	Memories   *prometheus.Desc

	DiskInfo  *prometheus.Desc
	DiskRead  *prometheus.Desc
	DiskWrite *prometheus.Desc

	NICInfo      *prometheus.Desc
	NICBandwidth *prometheus.Desc
	NICReceive   *prometheus.Desc
	NICSend      *prometheus.Desc

	MaintenanceScheduled *prometheus.Desc
	MaintenanceInfo      *prometheus.Desc
	MaintenanceStartTime *prometheus.Desc
	MaintenanceEndTime   *prometheus.Desc
}

// NewServerCollector returns a new ServerCollector.
func NewServerCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client iaas.ServerClient) *ServerCollector {
	errors.WithLabelValues("server").Add(0)

	serverLabels := []string{"id", "name", "zone"}
	serverInfoLabels := append(serverLabels, "cpus", "disks", "nics", "memories", "host", "tags", "description")
	diskLabels := append(serverLabels, "disk_id", "disk_name", "index")
	diskInfoLabels := append(diskLabels, "plan", "interface", "size", "tags", "description", "storage_id", "storage_generation", "storage_class")
	nicLabels := append(serverLabels, "interface_id", "index")
	nicInfoLabels := append(nicLabels, "upstream_type", "upstream_id", "upstream_name")
	maintenanceInfoLabel := append(serverLabels, "info_url", "info_title", "description", "start_date", "end_date")

	return &ServerCollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,
		Up: prometheus.NewDesc(
			"sakuracloud_server_up",
			"If 1 the server is up and running, 0 otherwise",
			serverLabels, nil,
		),
		ServerInfo: prometheus.NewDesc(
			"sakuracloud_server_info",
			"A metric with a constant '1' value labeled by server information",
			serverInfoLabels, nil,
		),
		CPUs: prometheus.NewDesc(
			"sakuracloud_server_cpus",
			"Number of server's vCPU cores",
			serverLabels, nil,
		),
		CPUTime: prometheus.NewDesc(
			"sakuracloud_server_cpu_time",
			"Server's CPU time(unit: ms)",
			serverLabels, nil,
		),
		Memories: prometheus.NewDesc(
			"sakuracloud_server_memories",
			"Size of server's memories(unit: GB)",
			serverLabels, nil,
		),
		DiskInfo: prometheus.NewDesc(
			"sakuracloud_server_disk_info",
			"A metric with a constant '1' value labeled by disk information",
			diskInfoLabels, nil,
		),
		DiskRead: prometheus.NewDesc(
			"sakuracloud_server_disk_read",
			"Disk's read bytes(unit: KBps)",
			diskLabels, nil,
		),
		DiskWrite: prometheus.NewDesc(
			"sakuracloud_server_disk_write",
			"Disk's write bytes(unit: KBps)",
			diskLabels, nil,
		),
		NICInfo: prometheus.NewDesc(
			"sakuracloud_server_nic_info",
			"A metric with a constant '1' value labeled by nic information",
			nicInfoLabels, nil,
		),
		NICBandwidth: prometheus.NewDesc(
			"sakuracloud_server_nic_bandwidth",
			"NIC's Bandwidth(unit: Mbps)",
			nicLabels, nil,
		),
		NICReceive: prometheus.NewDesc(
			"sakuracloud_server_nic_receive",
			"NIC's receive bytes(unit: Kbps)",
			nicLabels, nil,
		),
		NICSend: prometheus.NewDesc(
			"sakuracloud_server_nic_send",
			"NIC's send bytes(unit: Kbps)",
			nicLabels, nil,
		),
		MaintenanceScheduled: prometheus.NewDesc(
			"sakuracloud_server_maintenance_scheduled",
			"If 1 the server has scheduled maintenance info, 0 otherwise",
			serverLabels, nil,
		),
		MaintenanceInfo: prometheus.NewDesc(
			"sakuracloud_server_maintenance_info",
			"A metric with a constant '1' value labeled by maintenance information",
			maintenanceInfoLabel, nil,
		),
		MaintenanceStartTime: prometheus.NewDesc(
			"sakuracloud_server_maintenance_start",
			"Scheduled maintenance start time in seconds since epoch (1970)",
			serverLabels, nil,
		),
		MaintenanceEndTime: prometheus.NewDesc(
			"sakuracloud_server_maintenance_end",
			"Scheduled maintenance end time in seconds since epoch (1970)",
			serverLabels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *ServerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.ServerInfo
	ch <- c.CPUs
	ch <- c.CPUTime
	ch <- c.Memories

	ch <- c.DiskInfo
	ch <- c.DiskRead
	ch <- c.DiskWrite

	ch <- c.NICInfo
	ch <- c.NICBandwidth
	ch <- c.NICReceive
	ch <- c.NICSend
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *ServerCollector) Collect(ch chan<- prometheus.Metric) {
	servers, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("server").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", "can't list servers",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(servers))

	for i := range servers {
		func(server *iaas.Server) {
			defer wg.Done()

			serverLabels := c.serverLabels(server)

			var up float64
			if server.InstanceStatus.IsUp() {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Up,
				prometheus.GaugeValue,
				up,
				serverLabels...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.ServerInfo,
				prometheus.GaugeValue,
				float64(1.0),
				c.serverInfoLabels(server)...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.CPUs,
				prometheus.GaugeValue,
				float64(server.GetCPU()),
				serverLabels...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.Memories,
				prometheus.GaugeValue,
				float64(server.GetMemoryGB()),
				serverLabels...,
			)

			// maintenance info
			var maintenanceScheduled float64
			if server.InstanceHostInfoURL != "" {
				maintenanceScheduled = 1.0
				wg.Add(1)
				go func() {
					c.collectMaintenanceInfo(ch, server)
					wg.Done()
				}()
			}
			ch <- prometheus.MustNewConstMetric(
				c.MaintenanceScheduled,
				prometheus.GaugeValue,
				maintenanceScheduled,
				serverLabels...,
			)

			wg.Add(len(server.Disks))
			for i := range server.Disks {
				go func(i int) {
					c.collectDiskInfo(ch, server, i)
					wg.Done()
				}(i)
			}

			for i := range server.Interfaces {
				ch <- prometheus.MustNewConstMetric(
					c.NICInfo,
					prometheus.GaugeValue,
					float64(1.0),
					c.nicInfoLabels(server, i)...,
				)

				bandwidth := float64(server.BandWidthAt(i))
				ch <- prometheus.MustNewConstMetric(
					c.NICBandwidth,
					prometheus.GaugeValue,
					bandwidth,
					c.nicLabels(server, i)...,
				)
			}

			if server.InstanceStatus.IsUp() {
				// collect metrics per resources under server
				now := time.Now()
				// CPU-TIME
				wg.Add(1)
				go func() {
					c.collectCPUTime(ch, server, now)
					wg.Done()
				}()

				// Disks
				wg.Add(len(server.Disks))
				for i := range server.Disks {
					go func(i int) {
						c.collectDiskMetrics(ch, server, i, now)
						wg.Done()
					}(i)
				}

				// NICs
				wg.Add(len(server.Interfaces))
				for i := range server.Interfaces {
					go func(i int) {
						c.collectNICMetrics(ch, server, i, now)
						wg.Done()
					}(i)
				}
			}
		}(servers[i])
	}

	wg.Wait()
}

func (c *ServerCollector) serverLabels(server *iaas.Server) []string {
	return []string{
		server.ID.String(),
		server.Name,
		server.ZoneName,
	}
}

func (c *ServerCollector) serverInfoLabels(server *iaas.Server) []string {
	labels := c.serverLabels(server)

	instanceHost := "-"
	if server.InstanceHostName != "" {
		instanceHost = server.InstanceHostName
	}

	// append host/tags/descriptions
	return append(labels,
		fmt.Sprintf("%d", server.GetCPU()),
		fmt.Sprintf("%d", len(server.Disks)),
		fmt.Sprintf("%d", len(server.Interfaces)),
		fmt.Sprintf("%d", server.GetMemoryGB()),
		instanceHost,
		flattenStringSlice(server.Tags),
		server.Description,
	)
}

func (c *ServerCollector) serverMaintenanceInfoLabels(server *iaas.Server, info *newsfeed.FeedItem) []string {
	labels := c.serverLabels(server)

	return append(labels,
		info.URL,
		info.Title,
		info.Description,
		fmt.Sprintf("%d", info.EventStart().Unix()),
		fmt.Sprintf("%d", info.EventEnd().Unix()),
	)
}

var diskPlanLabels = map[types.ID]string{
	types.DiskPlans.HDD: "hdd",
	types.DiskPlans.SSD: "ssd",
}

func (c *ServerCollector) diskLabels(server *iaas.Server, index int) []string {
	if len(server.Disks) <= index {
		return nil
	}
	disk := server.Disks[index]
	return []string{
		server.ID.String(),
		server.Name,
		server.ZoneName,
		disk.ID.String(),
		disk.Name,
		fmt.Sprintf("%d", index),
	}
}

func (c *ServerCollector) collectDiskInfo(ch chan<- prometheus.Metric, server *iaas.Server, index int) {
	if len(server.Disks) <= index {
		return
	}
	labels := c.diskLabels(server, index)

	disk, err := c.client.ReadDisk(c.ctx, server.ZoneName, server.Disks[index].ID)
	if err != nil {
		c.errors.WithLabelValues("server").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", fmt.Sprintf("can't get server connected disk info: ID=%d, DiskID=%d", server.ID, server.Disks[index].ID),
			"err", err,
		)
		return
	}
	if disk == nil {
		return
	}

	var storageID, storageGeneration, storageClass string
	if disk.Storage != nil {
		storageID = disk.Storage.ID.String()
		storageGeneration = fmt.Sprintf("%d", disk.Storage.Generation)
		storageClass = disk.Storage.Class
	}

	labels = append(labels,
		diskPlanLabels[disk.DiskPlanID],
		string(disk.Connection),
		fmt.Sprintf("%d", disk.GetSizeGB()),
		flattenStringSlice(disk.Tags),
		disk.Description,
		storageID,
		storageGeneration,
		storageClass,
	)

	ch <- prometheus.MustNewConstMetric(
		c.DiskInfo,
		prometheus.GaugeValue,
		float64(1.0),
		labels...,
	)
}

func (c *ServerCollector) nicLabels(server *iaas.Server, index int) []string {
	if len(server.Interfaces) <= index {
		return nil
	}

	return []string{
		server.ID.String(),
		server.Name,
		server.ZoneName,
		server.Interfaces[index].ID.String(),
		fmt.Sprintf("%d", index),
	}
}

func (c *ServerCollector) nicInfoLabels(server *iaas.Server, index int) []string {
	if len(server.Interfaces) <= index {
		return nil
	}
	labels := c.nicLabels(server, index)

	upstreamType := server.Interfaces[index].UpstreamType.String()
	upstreamID := fmt.Sprintf("%d", server.Interfaces[index].SwitchID)
	if upstreamID == "-1" {
		upstreamID = ""
	}
	upstreamName := server.Interfaces[index].SwitchName

	return append(labels,
		upstreamType,
		upstreamID,
		upstreamName,
	)
}

func (c *ServerCollector) collectCPUTime(ch chan<- prometheus.Metric, server *iaas.Server, now time.Time) {
	values, err := c.client.MonitorCPU(c.ctx, server.ZoneName, server.ID, now)
	if err != nil {
		c.errors.WithLabelValues("server").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", fmt.Sprintf("can't get server's CPU-TIME: ID=%d", server.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	m := prometheus.MustNewConstMetric(
		c.CPUTime,
		prometheus.GaugeValue,
		values.CPUTime*1000,
		c.serverLabels(server)...,
	)

	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *ServerCollector) collectDiskMetrics(ch chan<- prometheus.Metric, server *iaas.Server, index int, now time.Time) {
	if len(server.Disks) <= index {
		return
	}
	disk := server.Disks[index]

	values, err := c.client.MonitorDisk(c.ctx, server.ZoneName, disk.ID, now)
	if err != nil {
		c.errors.WithLabelValues("server").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", fmt.Sprintf("can't get disk's metrics: ServerID=%d, DiskID=%d", server.ID, disk.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	read := values.Read
	if read > 0 {
		read = read / 1024
	}
	m := prometheus.MustNewConstMetric(
		c.DiskRead,
		prometheus.GaugeValue,
		read,
		c.diskLabels(server, index)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	write := values.Write
	if write > 0 {
		write = write / 1024
	}
	m = prometheus.MustNewConstMetric(
		c.DiskWrite,
		prometheus.GaugeValue,
		write,
		c.diskLabels(server, index)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *ServerCollector) collectNICMetrics(ch chan<- prometheus.Metric, server *iaas.Server, index int, now time.Time) {
	if len(server.Interfaces) <= index {
		return
	}
	nic := server.Interfaces[index]

	values, err := c.client.MonitorNIC(c.ctx, server.ZoneName, nic.ID, now)
	if err != nil {
		c.errors.WithLabelValues("server").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", fmt.Sprintf("can't get nic's metrics: ServerID=%d,NICID=%d", server.ID, nic.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	receive := values.Receive
	if receive > 0 {
		receive = receive * 8 / 1000
	}
	m := prometheus.MustNewConstMetric(
		c.NICReceive,
		prometheus.GaugeValue,
		receive,
		c.nicLabels(server, index)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	send := values.Send
	if send > 0 {
		send = send * 8 / 1000
	}
	m = prometheus.MustNewConstMetric(
		c.NICSend,
		prometheus.GaugeValue,
		send,
		c.nicLabels(server, index)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *ServerCollector) collectMaintenanceInfo(ch chan<- prometheus.Metric, server *iaas.Server) {
	if server.InstanceHostInfoURL == "" {
		return
	}
	info, err := c.client.MaintenanceInfo(server.InstanceHostInfoURL)
	if err != nil {
		c.errors.WithLabelValues("server").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", fmt.Sprintf("can't get server's maintenance info: ServerID=%d", server.ID),
			"err", err,
		)
		return
	}

	infoLabels := c.serverMaintenanceInfoLabels(server, info)

	// info
	ch <- prometheus.MustNewConstMetric(
		c.MaintenanceInfo,
		prometheus.GaugeValue,
		1.0,
		infoLabels...,
	)
	// start
	ch <- prometheus.MustNewConstMetric(
		c.MaintenanceStartTime,
		prometheus.GaugeValue,
		float64(info.EventStart().Unix()),
		c.serverLabels(server)...,
	)
	// end
	ch <- prometheus.MustNewConstMetric(
		c.MaintenanceEndTime,
		prometheus.GaugeValue,
		float64(info.EventEnd().Unix()),
		c.serverLabels(server)...,
	)
}
