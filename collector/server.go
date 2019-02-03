package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/sacloud/sakuracloud_exporter/iaas"
)

// ServerCollector collects metrics about all servers.
type ServerCollector struct {
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.ServerClient

	Up         *prometheus.Desc
	ServerInfo *prometheus.Desc
	CPUTime    *prometheus.Desc

	DiskInfo  *prometheus.Desc
	DiskRead  *prometheus.Desc
	DiskWrite *prometheus.Desc

	NICInfo      *prometheus.Desc
	NICBandwidth *prometheus.Desc
	NICReceive   *prometheus.Desc
	NICSend      *prometheus.Desc
}

// NewServerCollector returns a new ServerCollector.
func NewServerCollector(logger log.Logger, errors *prometheus.CounterVec, client iaas.ServerClient) *ServerCollector {
	errors.WithLabelValues("server").Add(0)

	serverLabels := []string{"id", "name", "zone"}
	serverInfoLabels := append(serverLabels, "cpus", "disks", "nics", "memories", "host", "tags", "description")
	diskLabels := append(serverLabels, "disk_id", "disk_name", "index")
	diskInfoLabels := append(diskLabels, "plan", "interface", "size", "tags", "description")
	nicLabels := append(serverLabels, "interface_id", "index")
	nicInfoLabels := append(nicLabels, "upstream_type", "upstream_id", "upstream_name")

	return &ServerCollector{
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
		CPUTime: prometheus.NewDesc(
			"sakuracloud_server_cpu_time",
			"Server's CPU time(unit: ms)",
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
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *ServerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.ServerInfo
	ch <- c.CPUTime

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
	servers, err := c.client.Find()
	if err != nil {
		c.errors.WithLabelValues("server").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list servers",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(servers))

	for i := range servers {
		go func(server *sacloud.Server) {
			defer wg.Done()

			serverLabels := c.serverLabels(server)

			var up float64
			if server.IsUp() {
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

			for i := range server.Disks {
				ch <- prometheus.MustNewConstMetric(
					c.DiskInfo,
					prometheus.GaugeValue,
					float64(1.0),
					c.diskInfoLabels(server, i)...,
				)
			}

			for i := range server.Interfaces {
				ch <- prometheus.MustNewConstMetric(
					c.NICInfo,
					prometheus.GaugeValue,
					float64(1.0),
					c.nicInfoLabels(server, i)...,
				)

				bandwidth := float64(server.BandwidthAt(i))
				ch <- prometheus.MustNewConstMetric(
					c.NICBandwidth,
					prometheus.GaugeValue,
					bandwidth,
					c.nicLabels(server, i)...,
				)
			}

			if server.IsUp() {
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

func (c *ServerCollector) serverLabels(server *sacloud.Server) []string {
	return []string{
		server.GetStrID(),
		server.Name,
		server.Zone.Name,
	}
}

func (c *ServerCollector) serverInfoLabels(server *sacloud.Server) []string {
	labels := c.serverLabels(server)

	instanceHost := "-"
	if server.Instance != nil {
		instanceHost = server.Instance.Host.Name
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

var diskPlanLabels = map[int64]string{
	int64(sacloud.DiskPlanHDDID): "hdd",
	int64(sacloud.DiskPlanSSDID): "ssd",
}

func (c *ServerCollector) diskLabels(server *sacloud.Server, index int) []string {
	if len(server.Disks) <= index {
		return nil
	}
	disk := server.Disks[index]
	return []string{
		server.GetStrID(),
		server.Name,
		server.GetZoneName(),
		disk.GetStrID(),
		disk.Name,
		fmt.Sprintf("%d", index),
	}
}

func (c *ServerCollector) diskInfoLabels(server *sacloud.Server, index int) []string {
	if len(server.Disks) <= index {
		return nil
	}
	labels := c.diskLabels(server, index)

	disk := server.Disks[index]

	return append(labels,
		diskPlanLabels[disk.GetPlanID()],
		string(disk.Connection),
		fmt.Sprintf("%d", disk.GetSizeGB()),
		flattenStringSlice(disk.Tags),
		disk.Description,
	)

}

func (c *ServerCollector) nicLabels(server *sacloud.Server, index int) []string {
	if len(server.Interfaces) <= index {
		return nil
	}

	return []string{
		server.GetStrID(),
		server.Name,
		server.GetZoneName(),
		server.Interfaces[index].GetStrID(),
		fmt.Sprintf("%d", index),
	}
}

func (c *ServerCollector) nicInfoLabels(server *sacloud.Server, index int) []string {
	if len(server.Interfaces) <= index {
		return nil
	}
	labels := c.nicLabels(server, index)

	upstreamType := server.Interfaces[index].UpstreamType().String()
	upstreamID := fmt.Sprintf("%d", server.SwitchIDAt(index))
	if upstreamID == "-1" {
		upstreamID = ""
	}
	upstreamName := server.SwitchNameAt(index)

	return append(labels,
		upstreamType,
		upstreamID,
		upstreamName,
	)
}

func (c *ServerCollector) collectCPUTime(ch chan<- prometheus.Metric, server *sacloud.Server, now time.Time) {
	values, err := c.client.MonitorCPU(server.GetZoneName(), server.ID, now)
	if err != nil {
		c.errors.WithLabelValues("server").Add(1)
		level.Warn(c.logger).Log(
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
		values.Value*1000,
		c.serverLabels(server)...,
	)

	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *ServerCollector) collectDiskMetrics(ch chan<- prometheus.Metric, server *sacloud.Server, index int, now time.Time) {

	if len(server.Disks) <= index {
		return
	}
	disk := server.Disks[index]

	values, err := c.client.MonitorDisk(server.GetZoneName(), disk.ID, now)
	if err != nil {
		c.errors.WithLabelValues("server").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get disk's metrics: ServerID=%d, DiskID=%d", server.ID, disk.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	if values.Read != nil {
		m := prometheus.MustNewConstMetric(
			c.DiskRead,
			prometheus.GaugeValue,
			values.Read.Value/1024,
			c.diskLabels(server, index)...,
		)
		ch <- prometheus.NewMetricWithTimestamp(values.Read.Time, m)
	}
	if values.Write != nil {
		m := prometheus.MustNewConstMetric(
			c.DiskWrite,
			prometheus.GaugeValue,
			values.Write.Value/1024,
			c.diskLabels(server, index)...,
		)
		ch <- prometheus.NewMetricWithTimestamp(values.Write.Time, m)
	}
}

func (c *ServerCollector) collectNICMetrics(ch chan<- prometheus.Metric, server *sacloud.Server, index int, now time.Time) {

	if len(server.Interfaces) <= index {
		return
	}
	nic := server.Interfaces[index]

	values, err := c.client.MonitorNIC(server.GetZoneName(), nic.ID, now)
	if err != nil {
		c.errors.WithLabelValues("server").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get nic's metrics: ServerID=%d,NICID=%d", server.ID, nic.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	if values.Receive != nil {
		m := prometheus.MustNewConstMetric(
			c.NICReceive,
			prometheus.GaugeValue,
			values.Receive.Value*8/1000,
			c.nicLabels(server, index)...,
		)
		ch <- prometheus.NewMetricWithTimestamp(values.Receive.Time, m)
	}
	if values.Send != nil {
		m := prometheus.MustNewConstMetric(
			c.NICSend,
			prometheus.GaugeValue,
			values.Send.Value*8/1000,
			c.nicLabels(server, index)...,
		)
		ch <- prometheus.NewMetricWithTimestamp(values.Send.Time, m)
	}
}
