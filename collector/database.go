// Copyright 2019-2023 The sakuracloud_exporter Authors
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
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/newsfeed"
	"github.com/sacloud/sakuracloud_exporter/platform"
)

// DatabaseCollector collects metrics about all databases.
type DatabaseCollector struct {
	ctx    context.Context
	logger log.Logger
	errors *prometheus.CounterVec
	client platform.DatabaseClient

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

	MaintenanceScheduled *prometheus.Desc
	MaintenanceInfo      *prometheus.Desc
	MaintenanceStartTime *prometheus.Desc
	MaintenanceEndTime   *prometheus.Desc
}

// NewDatabaseCollector returns a new DatabaseCollector.
func NewDatabaseCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client platform.DatabaseClient) *DatabaseCollector {
	errors.WithLabelValues("database").Add(0)

	databaseLabels := []string{"id", "name", "zone"}
	databaseInfoLabels := append(databaseLabels,
		"plan", "host",
		"database_type", "database_revision", "database_version",
		"web_ui", "replication_enabled", "replication_role", "tags", "description")

	nicInfoLabels := append(databaseLabels, "upstream_type", "upstream_id", "upstream_name", "ipaddress", "nw_mask_len", "gateway")

	return &DatabaseCollector{
		ctx:    ctx,
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
		MaintenanceScheduled: prometheus.NewDesc(
			"sakuracloud_database_maintenance_scheduled",
			"If 1 the database has scheduled maintenance info, 0 otherwise",
			databaseLabels, nil,
		),
		MaintenanceInfo: prometheus.NewDesc(
			"sakuracloud_database_maintenance_info",
			"A metric with a constant '1' value labeled by maintenance information",
			append(databaseLabels, "info_url", "info_title", "description", "start_date", "end_date"), nil,
		),
		MaintenanceStartTime: prometheus.NewDesc(
			"sakuracloud_database_maintenance_start",
			"Scheduled maintenance start time in seconds since epoch (1970)",
			databaseLabels, nil,
		),
		MaintenanceEndTime: prometheus.NewDesc(
			"sakuracloud_database_maintenance_end",
			"Scheduled maintenance end time in seconds since epoch (1970)",
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

	ch <- c.MaintenanceScheduled
	ch <- c.MaintenanceInfo
	ch <- c.MaintenanceStartTime
	ch <- c.MaintenanceEndTime
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *DatabaseCollector) Collect(ch chan<- prometheus.Metric) {
	databases, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("database").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", "can't list databases",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(databases))

	for i := range databases {
		func(database *platform.Database) {
			defer wg.Done()

			databaseLabels := c.databaseLabels(database)

			var up float64
			if database.InstanceStatus.IsUp() {
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

			if database.Availability.IsAvailable() && database.InstanceStatus.IsUp() {
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

				// maintenance info
				var maintenanceScheduled float64
				if database.InstanceHostInfoURL != "" {
					maintenanceScheduled = 1.0
					wg.Add(1)
					go func() {
						c.collectMaintenanceInfo(ch, database)
						wg.Done()
					}()
				}
				ch <- prometheus.MustNewConstMetric(
					c.MaintenanceScheduled,
					prometheus.GaugeValue,
					maintenanceScheduled,
					databaseLabels...,
				)
			}
		}(databases[i])
	}

	wg.Wait()
}

func (c *DatabaseCollector) databaseLabels(database *platform.Database) []string {
	return []string{
		database.ID.String(),
		database.Name,
		database.ZoneName,
	}
}

var databasePlanLabels = map[types.ID]string{
	types.DatabasePlans.DB10GB:  "10GB",
	types.DatabasePlans.DB30GB:  "30GB",
	types.DatabasePlans.DB90GB:  "90GB",
	types.DatabasePlans.DB240GB: "240GB",
	types.DatabasePlans.DB500GB: "500GB",
	types.DatabasePlans.DB1TB:   "1TB",
}

func (c *DatabaseCollector) databaseInfoLabels(database *platform.Database) []string {
	labels := c.databaseLabels(database)

	instanceHost := "-"
	if database.InstanceHostName != "" {
		instanceHost = database.InstanceHostName
	}

	replEnabled := "0"
	replRole := ""
	if database.ReplicationSetting != nil {
		replEnabled = "1"
		if database.ReplicationSetting.Model == types.DatabaseReplicationModels.MasterSlave {
			replRole = "master"
		} else {
			replRole = "slave"
		}
	}

	return append(labels,
		databasePlanLabels[database.PlanID],
		instanceHost,
		database.Conf.DatabaseName,
		database.Conf.DatabaseRevision,
		database.Conf.DatabaseVersion,
		"", // TODO libsacloud v2 doesn't support WebUI URL
		replEnabled,
		replRole,
		flattenStringSlice(database.Tags),
		database.Description,
	)
}

func (c *DatabaseCollector) nicInfoLabels(database *platform.Database) []string {
	labels := c.databaseLabels(database)

	var upstreamType, upstreamID, upstreamName string

	if len(database.Interfaces) > 0 {
		nic := database.Interfaces[0]

		upstreamType = nic.UpstreamType.String()
		if !nic.SwitchID.IsEmpty() {
			upstreamID = nic.SwitchID.String()
			upstreamName = nic.SwitchName
		}
	}

	nwMaskLen := database.NetworkMaskLen
	strMaskLen := ""
	if nwMaskLen > 0 {
		strMaskLen = fmt.Sprintf("%d", nwMaskLen)
	}

	return append(labels,
		upstreamType,
		upstreamID,
		upstreamName,
		database.IPAddresses[0],
		strMaskLen,
		database.DefaultRoute,
	)
}

func (c *DatabaseCollector) collectCPUTime(ch chan<- prometheus.Metric, database *platform.Database, now time.Time) {
	values, err := c.client.MonitorCPU(c.ctx, database.ZoneName, database.ID, now)
	if err != nil {
		c.errors.WithLabelValues("database").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", fmt.Sprintf("can't get database's cpu time: DatabaseID=%d", database.ID),
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
		c.databaseLabels(database)...,
	)

	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *DatabaseCollector) collectDiskMetrics(ch chan<- prometheus.Metric, database *platform.Database, now time.Time) {
	values, err := c.client.MonitorDisk(c.ctx, database.ZoneName, database.ID, now)
	if err != nil {
		c.errors.WithLabelValues("database").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", fmt.Sprintf("can't get disk's metrics: DatabaseID=%d", database.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	m := prometheus.MustNewConstMetric(
		c.DiskRead,
		prometheus.GaugeValue,
		values.Read/1024,
		c.databaseLabels(database)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
	m = prometheus.MustNewConstMetric(
		c.DiskWrite,
		prometheus.GaugeValue,
		values.Write/1024,
		c.databaseLabels(database)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *DatabaseCollector) collectNICMetrics(ch chan<- prometheus.Metric, database *platform.Database, now time.Time) {
	values, err := c.client.MonitorNIC(c.ctx, database.ZoneName, database.ID, now)
	if err != nil {
		c.errors.WithLabelValues("database").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", fmt.Sprintf("can't get database's NIC metrics: DatabaseID=%d", database.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	m := prometheus.MustNewConstMetric(
		c.NICReceive,
		prometheus.GaugeValue,
		values.Receive*8/1000,
		c.databaseLabels(database)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	m = prometheus.MustNewConstMetric(
		c.NICSend,
		prometheus.GaugeValue,
		values.Send*8/1000,
		c.databaseLabels(database)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *DatabaseCollector) collectDatabaseMetrics(ch chan<- prometheus.Metric, database *platform.Database, now time.Time) {
	values, err := c.client.MonitorDatabase(c.ctx, database.ZoneName, database.ID, now)
	if err != nil {
		c.errors.WithLabelValues("database").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", fmt.Sprintf("can't get database's system metrics: DatabaseID=%d", database.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	labels := c.databaseLabels(database)
	totalMemorySize := values.TotalMemorySize
	if totalMemorySize > 0 {
		totalMemorySize = totalMemorySize / 1024 / 1024
	}
	m := prometheus.MustNewConstMetric(
		c.MemoryTotal,
		prometheus.GaugeValue,
		totalMemorySize,
		labels...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	usedMemorySize := values.UsedMemorySize
	if usedMemorySize > 0 {
		usedMemorySize = usedMemorySize / 1024 / 1024
	}
	m = prometheus.MustNewConstMetric(
		c.MemoryUsed,
		prometheus.GaugeValue,
		usedMemorySize,
		labels...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	totalDisk1Size := values.TotalDisk1Size
	if totalDisk1Size > 0 {
		totalDisk1Size = totalDisk1Size / 1024 / 1024
	}
	m = prometheus.MustNewConstMetric(
		c.SystemDiskTotal,
		prometheus.GaugeValue,
		totalDisk1Size,
		labels...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	usedDisk1Size := values.UsedDisk1Size
	if usedDisk1Size > 0 {
		usedDisk1Size = usedDisk1Size / 1024 / 1024
	}
	m = prometheus.MustNewConstMetric(
		c.SystemDiskUsed,
		prometheus.GaugeValue,
		usedDisk1Size,
		labels...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	totalDisk2Size := values.TotalDisk2Size
	if totalDisk2Size > 0 {
		totalDisk2Size = totalDisk2Size / 1024 / 1024
	}
	m = prometheus.MustNewConstMetric(
		c.BackupDiskTotal,
		prometheus.GaugeValue,
		totalDisk2Size,
		labels...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	usedDisk2Size := values.UsedDisk2Size
	if usedDisk2Size > 0 {
		usedDisk2Size = usedDisk2Size / 1024 / 1024
	}
	m = prometheus.MustNewConstMetric(
		c.BackupDiskUsed,
		prometheus.GaugeValue,
		usedDisk2Size,
		labels...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	binlogUsed := values.BinlogUsedSizeKiB
	if binlogUsed > 0 {
		binlogUsed = binlogUsed / 1024 / 1024
	}
	m = prometheus.MustNewConstMetric(
		c.BinlogUsed,
		prometheus.GaugeValue,
		binlogUsed,
		c.databaseLabels(database)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	m = prometheus.MustNewConstMetric(
		c.ReplicationDelay,
		prometheus.GaugeValue,
		values.DelayTimeSec,
		c.databaseLabels(database)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *DatabaseCollector) maintenanceInfoLabels(resource *platform.Database, info *newsfeed.FeedItem) []string {
	labels := c.databaseLabels(resource)

	return append(labels,
		info.URL,
		info.Title,
		info.Description,
		fmt.Sprintf("%d", info.EventStart().Unix()),
		fmt.Sprintf("%d", info.EventEnd().Unix()),
	)
}

func (c *DatabaseCollector) collectMaintenanceInfo(ch chan<- prometheus.Metric, resource *platform.Database) {
	if resource.InstanceHostInfoURL == "" {
		return
	}
	info, err := c.client.MaintenanceInfo(resource.InstanceHostInfoURL)
	if err != nil {
		c.errors.WithLabelValues("database").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", fmt.Sprintf("can't get database's maintenance info: ID=%d", resource.ID),
			"err", err,
		)
		return
	}

	infoLabels := c.maintenanceInfoLabels(resource, info)

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
		c.databaseLabels(resource)...,
	)
	// end
	ch <- prometheus.MustNewConstMetric(
		c.MaintenanceEndTime,
		prometheus.GaugeValue,
		float64(info.EventEnd().Unix()),
		c.databaseLabels(resource)...,
	)
}
