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

// DatabaseCollector collects metrics about all databases.
type DatabaseCollector struct {
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.DatabaseClient

	Up               *prometheus.Desc
	DatabaseInfo     *prometheus.Desc
	CPUTime          *prometheus.Desc
	MemoryUsed       *prometheus.Desc
	MemoryTotal      *prometheus.Desc
	NICInfo          *prometheus.Desc
	NICReceive       *prometheus.Desc
	NICSend          *prometheus.Desc
	SystemDiskUsed   *prometheus.Desc
	SystemDiskTotal  *prometheus.Desc
	BackupDiskUsed   *prometheus.Desc
	BackupDiskTotal  *prometheus.Desc
	BinlogUsed       *prometheus.Desc
	DiskRead         *prometheus.Desc
	DiskWrite        *prometheus.Desc
	ReplicationDelay *prometheus.Desc
}

// NewDatabaseCollector returns a new DatabaseCollector.
func NewDatabaseCollector(logger log.Logger, errors *prometheus.CounterVec, client iaas.DatabaseClient) *DatabaseCollector {
	errors.WithLabelValues("database").Add(0)

	databaseLabels := []string{"id", "name", "zone"}
	databaseInfoLabels := append(databaseLabels,
		"plan", "host",
		"database_type", "database_revision", "database_version",
		"web_ui", "replication_enabled", "replication_role", "tags", "description")

	nicInfoLabels := append(databaseLabels, "upstream_type", "upstream_id", "upstream_name", "ipaddress", "nw_mask_len", "gateway")

	return &DatabaseCollector{
		logger: logger,
		errors: errors,
		client: client,
		Up: prometheus.NewDesc(
			"sakuracloud_database_up",
			"If 1 the database is up and running, 0 otherwise",
			databaseLabels, nil,
		),
		DatabaseInfo: prometheus.NewDesc(
			"sakuracloud_database_info",
			"A metric with a constant '1' value labeled by database information",
			databaseInfoLabels, nil,
		),
		CPUTime: prometheus.NewDesc(
			"sakuracloud_database_cpu_time",
			"Database's CPU time(unit:ms)",
			databaseLabels, nil,
		),
		MemoryUsed: prometheus.NewDesc(
			"sakuracloud_database_memory_used",
			"Database's used memory size(unit:GB)",
			databaseLabels, nil,
		),
		MemoryTotal: prometheus.NewDesc(
			"sakuracloud_database_memory_total",
			"Database's total memory size(unit:GB)",
			databaseLabels, nil,
		),
		NICInfo: prometheus.NewDesc(
			"sakuracloud_database_nic_info",
			"A metric with a constant '1' value labeled by nic information",
			nicInfoLabels, nil,
		),
		NICReceive: prometheus.NewDesc(
			"sakuracloud_database_nic_receive",
			"NIC's receive bytes(unit: Kbps)",
			databaseLabels, nil,
		),
		NICSend: prometheus.NewDesc(
			"sakuracloud_database_nic_send",
			"NIC's send bytes(unit: Kbps)",
			databaseLabels, nil,
		),
		SystemDiskUsed: prometheus.NewDesc(
			"sakuracloud_database_disk_system_used",
			"Database's used system-disk size(unit:GB)",
			databaseLabels, nil,
		),
		SystemDiskTotal: prometheus.NewDesc(
			"sakuracloud_database_disk_system_total",
			"Database's total system-disk size(unit:GB)",
			databaseLabels, nil,
		),
		BackupDiskUsed: prometheus.NewDesc(
			"sakuracloud_database_disk_backup_used",
			"Database's used backup-disk size(unit:GB)",
			databaseLabels, nil,
		),
		BackupDiskTotal: prometheus.NewDesc(
			"sakuracloud_database_disk_backup_total",
			"Database's total backup-disk size(unit:GB)",
			databaseLabels, nil,
		),
		BinlogUsed: prometheus.NewDesc(
			"sakuracloud_database_binlog_used",
			"Database's used binlog size(unit:GB)",
			databaseLabels, nil,
		),
		DiskRead: prometheus.NewDesc(
			"sakuracloud_database_disk_read",
			"Disk's read bytes(unit: KBps)",
			databaseLabels, nil,
		),
		DiskWrite: prometheus.NewDesc(
			"sakuracloud_database_disk_write",
			"Disk's write bytes(unit: KBps)",
			databaseLabels, nil,
		),
		ReplicationDelay: prometheus.NewDesc(
			"sakuracloud_database_replication_delay",
			"Replication delay time(unit:second)",
			databaseLabels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *DatabaseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.DatabaseInfo
	ch <- c.CPUTime
	ch <- c.MemoryUsed
	ch <- c.MemoryTotal
	ch <- c.NICInfo
	ch <- c.NICReceive
	ch <- c.NICSend
	ch <- c.SystemDiskUsed
	ch <- c.SystemDiskTotal
	ch <- c.BackupDiskUsed
	ch <- c.BackupDiskTotal
	ch <- c.BinlogUsed
	ch <- c.DiskRead
	ch <- c.DiskWrite
	ch <- c.ReplicationDelay
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *DatabaseCollector) Collect(ch chan<- prometheus.Metric) {
	databases, err := c.client.Find()
	if err != nil {
		c.errors.WithLabelValues("database").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list databases",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(databases))

	for i := range databases {
		go func(database *iaas.Database) {
			defer wg.Done()

			databaseLabels := c.databaseLabels(database)

			var up float64
			if database.IsUp() {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Up,
				prometheus.GaugeValue,
				up,
				databaseLabels...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.DatabaseInfo,
				prometheus.GaugeValue,
				float64(1.0),
				c.databaseInfoLabels(database)...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.NICInfo,
				prometheus.GaugeValue,
				float64(1.0),
				c.nicInfoLabels(database)...,
			)

			if database.IsUp() {
				now := time.Now()

				// system info
				wg.Add(1)
				go func() {
					c.collectDatabaseMetrics(ch, database, now)
					wg.Done()
				}()

				// cpu-time
				wg.Add(1)
				go func() {
					c.collectCPUTime(ch, database, now)
					wg.Done()
				}()

				// Disk read/write
				wg.Add(1)
				go func() {
					c.collectDiskMetrics(ch, database, now)
					wg.Done()
				}()

				// NICs
				wg.Add(1)
				go func() {
					c.collectNICMetrics(ch, database, now)
					wg.Done()
				}()
			}

		}(databases[i])
	}

	wg.Wait()
}

func (c *DatabaseCollector) databaseLabels(database *iaas.Database) []string {
	return []string{
		database.GetStrID(),
		database.Name,
		database.ZoneName,
	}
}

var databasePlanLabels = map[int64]string{
	int64(sacloud.DatabasePlan10G):  "10GB",
	int64(sacloud.DatabasePlan30G):  "30GB",
	int64(sacloud.DatabasePlan90G):  "90GB",
	int64(sacloud.DatabasePlan240G): "240GB",
	int64(sacloud.DatabasePlan500G): "500GB",
	int64(sacloud.DatabasePlan1T):   "1TB",
}

func (c *DatabaseCollector) databaseInfoLabels(database *iaas.Database) []string {
	labels := c.databaseLabels(database)

	instanceHost := "-"
	if database.Instance != nil {
		instanceHost = database.Instance.Host.Name
	}

	replEnabled := "0"
	replRole := ""
	if database.IsReplicationEnabled() {
		replEnabled = "1"
		if database.IsReplicationMaster() {
			replRole = "master"
		} else {
			replRole = "slave"
		}
	}

	return append(labels,
		databasePlanLabels[database.Plan.ID],
		instanceHost,
		database.DatabaseName(),
		database.DatabaseRevision(),
		database.DatabaseVersion(),
		database.WebUIAddress(),
		replEnabled,
		replRole,
		flattenStringSlice(database.Tags),
		database.Description,
	)
}

func (c *DatabaseCollector) nicInfoLabels(database *iaas.Database) []string {
	labels := c.databaseLabels(database)

	var upstreamType, upstreamID, upstreamName string

	if len(database.Interfaces) > 0 {
		nic := database.GetFirstInterface()

		upstreamType = nic.UpstreamType().String()
		if nic.Switch != nil {
			upstreamID = nic.Switch.GetStrID()
			upstreamName = nic.Switch.Name
		}
	}

	nwMaskLen := database.NetworkMaskLen()
	strMaskLen := ""
	if nwMaskLen > 0 {
		strMaskLen = fmt.Sprintf("%d", nwMaskLen)
	}

	return append(labels,
		upstreamType,
		upstreamID,
		upstreamName,
		database.IPAddress(),
		strMaskLen,
		database.DefaultRoute(),
	)
}

func (c *DatabaseCollector) collectCPUTime(ch chan<- prometheus.Metric, database *iaas.Database, now time.Time) {

	values, err := c.client.MonitorCPU(database.ZoneName, database.ID, now)
	if err != nil {
		c.errors.WithLabelValues("database").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get database's cpu time: DatabaseID=%d", database.ID),
			"err", err,
		)
		return
	}
	if len(values) == 0 {
		return
	}

	for _, v := range values {
		m := prometheus.MustNewConstMetric(
			c.CPUTime,
			prometheus.GaugeValue,
			v.Value*1000,
			c.databaseLabels(database)...,
		)

		ch <- prometheus.NewMetricWithTimestamp(v.Time, m)
	}
}

func (c *DatabaseCollector) collectDiskMetrics(ch chan<- prometheus.Metric, database *iaas.Database, now time.Time) {

	values, err := c.client.MonitorDisk(database.ZoneName, database.ID, now)
	if err != nil {
		c.errors.WithLabelValues("database").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get disk's metrics: DatabaseID=%d", database.ID),
			"err", err,
		)
		return
	}
	if len(values) == 0 {
		return
	}

	for _, v := range values {
		if v.Read != nil {
			m := prometheus.MustNewConstMetric(
				c.DiskRead,
				prometheus.GaugeValue,
				v.Read.Value/1024,
				c.databaseLabels(database)...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.Read.Time, m)
		}
		if v.Write != nil {
			m := prometheus.MustNewConstMetric(
				c.DiskWrite,
				prometheus.GaugeValue,
				v.Write.Value/1024,
				c.databaseLabels(database)...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.Write.Time, m)
		}
	}
}

func (c *DatabaseCollector) collectNICMetrics(ch chan<- prometheus.Metric, database *iaas.Database, now time.Time) {

	values, err := c.client.MonitorNIC(database.ZoneName, database.ID, now)
	if err != nil {
		c.errors.WithLabelValues("database").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get database's NIC metrics: DatabaseID=%d", database.ID),
			"err", err,
		)
		return
	}
	if len(values) == 0 {
		return
	}

	for _, v := range values {
		if v.Receive != nil {
			m := prometheus.MustNewConstMetric(
				c.NICReceive,
				prometheus.GaugeValue,
				v.Receive.Value*8/1000,
				c.databaseLabels(database)...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.Receive.Time, m)
		}
		if v.Send != nil {
			m := prometheus.MustNewConstMetric(
				c.NICSend,
				prometheus.GaugeValue,
				v.Send.Value*8/1000,
				c.databaseLabels(database)...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.Send.Time, m)
		}
	}
}

func (c *DatabaseCollector) collectDatabaseMetrics(ch chan<- prometheus.Metric, database *iaas.Database, now time.Time) {

	values, err := c.client.MonitorDatabase(database.ZoneName, database.ID, now)
	if err != nil {
		c.errors.WithLabelValues("database").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get database's system metrics: DatabaseID=%d", database.ID),
			"err", err,
		)
		return
	}
	if len(values) == 0 {
		return
	}

	labels := c.databaseLabels(database)
	for _, v := range values {
		if v.TotalMemorySize != nil && v.TotalMemorySize.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.MemoryTotal,
				prometheus.GaugeValue,
				v.TotalMemorySize.Value/1024/1024,
				labels...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.TotalMemorySize.Time, m)
		}
		if v.UsedMemorySize != nil && v.UsedMemorySize.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.MemoryUsed,
				prometheus.GaugeValue,
				v.UsedMemorySize.Value/1024/1024,
				labels...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.UsedMemorySize.Time, m)
		}
		if v.TotalDisk1Size != nil && v.TotalDisk1Size.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.SystemDiskTotal,
				prometheus.GaugeValue,
				v.TotalDisk1Size.Value/1024/1024,
				labels...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.TotalDisk1Size.Time, m)
		}
		if v.UsedDisk1Size != nil && v.UsedDisk1Size.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.SystemDiskUsed,
				prometheus.GaugeValue,
				v.UsedDisk1Size.Value/1024/1024,
				labels...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.UsedDisk1Size.Time, m)
		}
		if v.TotalDisk2Size != nil && v.TotalDisk2Size.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.BackupDiskTotal,
				prometheus.GaugeValue,
				v.TotalDisk2Size.Value/1024/1024,
				labels...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.TotalDisk2Size.Time, m)
		}
		if v.UsedDisk2Size != nil && v.UsedDisk2Size.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.BackupDiskUsed,
				prometheus.GaugeValue,
				v.UsedDisk2Size.Value/1024/1024,
				labels...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.UsedDisk2Size.Time, m)
		}
		if v.BinlogUsedSizeKiB != nil && v.BinlogUsedSizeKiB.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.BinlogUsed,
				prometheus.GaugeValue,
				v.BinlogUsedSizeKiB.Value/1024/1024,
				c.databaseLabels(database)...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.BinlogUsedSizeKiB.Time, m)
		}
		if v.DelayTimeSec != nil && v.DelayTimeSec.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.ReplicationDelay,
				prometheus.GaugeValue,
				v.DelayTimeSec.Value,
				c.databaseLabels(database)...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.DelayTimeSec.Time, m)
		}
	}
}
