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
	"github.com/sacloud/packages-go/newsfeed"
	"github.com/sacloud/sakuracloud_exporter/platform"
)

// NFSCollector collects metrics about all nfss.
type NFSCollector struct {
	ctx    context.Context
	logger log.Logger
	errors *prometheus.CounterVec
	client platform.NFSClient

	Up      *prometheus.Desc
	NFSInfo *prometheus.Desc

	DiskFree *prometheus.Desc

	NICInfo    *prometheus.Desc
	NICReceive *prometheus.Desc
	NICSend    *prometheus.Desc

	MaintenanceScheduled *prometheus.Desc
	MaintenanceInfo      *prometheus.Desc
	MaintenanceStartTime *prometheus.Desc
	MaintenanceEndTime   *prometheus.Desc
}

// NewNFSCollector returns a new NFSCollector.
func NewNFSCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client platform.NFSClient) *NFSCollector {
	errors.WithLabelValues("nfs").Add(0)

	nfsLabels := []string{"id", "name", "zone"}
	nfsInfoLabels := append(nfsLabels, "plan", "size", "host", "tags", "description")
	nicInfoLabels := append(nfsLabels, "upstream_id", "upstream_name", "ipaddress", "nw_mask_len", "gateway")

	return &NFSCollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,
		Up: prometheus.NewDesc(
			"sakuracloud_nfs_up",
			"If 1 the nfs is up and running, 0 otherwise",
			nfsLabels, nil,
		),
		NFSInfo: prometheus.NewDesc(
			"sakuracloud_nfs_info",
			"A metric with a constant '1' value labeled by nfs information",
			nfsInfoLabels, nil,
		),
		DiskFree: prometheus.NewDesc(
			"sakuracloud_nfs_free_disk_size",
			"NFS's Free Disk Size(unit: GB)",
			nfsLabels, nil,
		),
		NICInfo: prometheus.NewDesc(
			"sakuracloud_nfs_nic_info",
			"A metric with a constant '1' value labeled by nic information",
			nicInfoLabels, nil,
		),
		NICReceive: prometheus.NewDesc(
			"sakuracloud_nfs_receive",
			"NIC's receive bytes(unit: Kbps)",
			nfsLabels, nil,
		),
		NICSend: prometheus.NewDesc(
			"sakuracloud_nfs_send",
			"NIC's send bytes(unit: Kbps)",
			nfsLabels, nil,
		),
		MaintenanceScheduled: prometheus.NewDesc(
			"sakuracloud_nfs_maintenance_scheduled",
			"If 1 the nfs has scheduled maintenance info, 0 otherwise",
			nfsLabels, nil,
		),
		MaintenanceInfo: prometheus.NewDesc(
			"sakuracloud_nfs_maintenance_info",
			"A metric with a constant '1' value labeled by maintenance information",
			append(nfsLabels, "info_url", "info_title", "description", "start_date", "end_date"), nil,
		),
		MaintenanceStartTime: prometheus.NewDesc(
			"sakuracloud_nfs_maintenance_start",
			"Scheduled maintenance start time in seconds since epoch (1970)",
			nfsLabels, nil,
		),
		MaintenanceEndTime: prometheus.NewDesc(
			"sakuracloud_nfs_maintenance_end",
			"Scheduled maintenance end time in seconds since epoch (1970)",
			nfsLabels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *NFSCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.NFSInfo
	ch <- c.DiskFree
	ch <- c.NICInfo
	ch <- c.NICReceive
	ch <- c.NICSend

	ch <- c.MaintenanceScheduled
	ch <- c.MaintenanceInfo
	ch <- c.MaintenanceStartTime
	ch <- c.MaintenanceEndTime
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *NFSCollector) Collect(ch chan<- prometheus.Metric) {
	nfss, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("nfs").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", "can't list nfs",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(nfss))

	for i := range nfss {
		func(nfs *platform.NFS) {
			defer wg.Done()

			nfsLabels := c.nfsLabels(nfs)

			var up float64
			if nfs.InstanceStatus.IsUp() {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Up,
				prometheus.GaugeValue,
				up,
				nfsLabels...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.NFSInfo,
				prometheus.GaugeValue,
				float64(1.0),
				c.nfsInfoLabels(nfs)...,
			)

			ch <- prometheus.MustNewConstMetric(
				c.NICInfo,
				prometheus.GaugeValue,
				float64(1.0),
				c.nicInfoLabels(nfs)...,
			)

			if nfs.Availability.IsAvailable() && nfs.InstanceStatus.IsUp() {
				now := time.Now()
				// Free disk size
				wg.Add(1)
				go func() {
					c.collectFreeDiskSize(ch, nfs, now)
					wg.Done()
				}()

				// NICs
				wg.Add(1)
				go func() {
					c.collectNICMetrics(ch, nfs, now)
					wg.Done()
				}()

				// maintenance info
				var maintenanceScheduled float64
				if nfs.InstanceHostInfoURL != "" {
					maintenanceScheduled = 1.0
					wg.Add(1)
					go func() {
						c.collectMaintenanceInfo(ch, nfs)
						wg.Done()
					}()
				}
				ch <- prometheus.MustNewConstMetric(
					c.MaintenanceScheduled,
					prometheus.GaugeValue,
					maintenanceScheduled,
					nfsLabels...,
				)
			}
		}(nfss[i])
	}

	wg.Wait()
}

func (c *NFSCollector) nfsLabels(nfs *platform.NFS) []string {
	return []string{
		nfs.ID.String(),
		nfs.Name,
		nfs.ZoneName,
	}
}

func (c *NFSCollector) nfsInfoLabels(nfs *platform.NFS) []string {
	labels := c.nfsLabels(nfs)

	instanceHost := "-"
	if nfs.InstanceHostName != "" {
		instanceHost = nfs.InstanceHostName
	}

	var plan string
	var size string
	if nfs.Plan != nil {
		plan = nfs.PlanName
		size = fmt.Sprintf("%d", nfs.Plan.Size)
	}

	return append(labels,
		plan,
		size,
		instanceHost,
		flattenStringSlice(nfs.Tags),
		nfs.Description,
	)
}

func (c *NFSCollector) nicInfoLabels(nfs *platform.NFS) []string {
	labels := c.nfsLabels(nfs)

	upstreamID := nfs.SwitchID.String()
	upstreamName := nfs.SwitchName

	ip := ""
	if len(nfs.IPAddresses) > 0 {
		ip = nfs.IPAddresses[0]
	}

	nwMaskLen := nfs.NetworkMaskLen
	strMaskLen := ""
	if nwMaskLen > 0 {
		strMaskLen = fmt.Sprintf("%d", nwMaskLen)
	}

	return append(labels,
		upstreamID,
		upstreamName,
		ip,
		strMaskLen,
		nfs.DefaultRoute,
	)
}

func (c *NFSCollector) collectFreeDiskSize(ch chan<- prometheus.Metric, nfs *platform.NFS, now time.Time) {
	values, err := c.client.MonitorFreeDiskSize(c.ctx, nfs.ZoneName, nfs.ID, now)
	if err != nil {
		c.errors.WithLabelValues("nfs").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", fmt.Sprintf("can't get disk's free size: NFSID=%d", nfs.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	v := values.FreeDiskSize
	if v > 0 {
		v = v / 1024 / 1024
	}
	m := prometheus.MustNewConstMetric(
		c.DiskFree,
		prometheus.GaugeValue,
		v,
		c.nfsLabels(nfs)...,
	)

	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *NFSCollector) collectNICMetrics(ch chan<- prometheus.Metric, nfs *platform.NFS, now time.Time) {
	values, err := c.client.MonitorNIC(c.ctx, nfs.ZoneName, nfs.ID, now)
	if err != nil {
		c.errors.WithLabelValues("nfs").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", fmt.Sprintf("can't get nfs's NIC metrics: NFSID=%d", nfs.ID),
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
		c.nfsLabels(nfs)...,
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
		c.nfsLabels(nfs)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *NFSCollector) maintenanceInfoLabels(resource *platform.NFS, info *newsfeed.FeedItem) []string {
	labels := c.nfsLabels(resource)

	return append(labels,
		info.URL,
		info.Title,
		info.Description,
		fmt.Sprintf("%d", info.EventStart().Unix()),
		fmt.Sprintf("%d", info.EventEnd().Unix()),
	)
}

func (c *NFSCollector) collectMaintenanceInfo(ch chan<- prometheus.Metric, resource *platform.NFS) {
	if resource.InstanceHostInfoURL == "" {
		return
	}
	info, err := c.client.MaintenanceInfo(resource.InstanceHostInfoURL)
	if err != nil {
		c.errors.WithLabelValues("nfs").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", fmt.Sprintf("can't get nfs's maintenance info: ID=%d", resource.ID),
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
		c.nfsLabels(resource)...,
	)
	// end
	ch <- prometheus.MustNewConstMetric(
		c.MaintenanceEndTime,
		prometheus.GaugeValue,
		float64(info.EventEnd().Unix()),
		c.nfsLabels(resource)...,
	)
}
