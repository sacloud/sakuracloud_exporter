// Copyright 2019-2022 The sakuracloud_exporter Authors
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
	"sort"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/sakuracloud_exporter/iaas"
)

// AutoBackupCollector collects metrics about all auto_backups.
type AutoBackupCollector struct {
	ctx    context.Context
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.AutoBackupClient

	Info *prometheus.Desc

	BackupCount    *prometheus.Desc
	LastBackupTime *prometheus.Desc
	BackupInfo     *prometheus.Desc
}

// NewAutoBackupCollector returns a new AutoBackupCollector.
func NewAutoBackupCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client iaas.AutoBackupClient) *AutoBackupCollector {
	errors.WithLabelValues("auto_backup").Add(0)

	labels := []string{"id", "name", "disk_id"}
	infoLabels := append(labels, "max_backup_num", "weekdays", "tags", "description")
	backupLabels := append(labels, "archive_id", "archive_name", "archive_tags", "archive_description")

	return &AutoBackupCollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,
		Info: prometheus.NewDesc(
			"sakuracloud_auto_backup_info",
			"A metric with a constant '1' value labeled by auto_backup information",
			infoLabels, nil,
		),
		BackupCount: prometheus.NewDesc(
			"sakuracloud_auto_backup_count",
			"A count of archives created by AutoBackup",
			labels, nil,
		),
		LastBackupTime: prometheus.NewDesc(
			"sakuracloud_auto_backup_last_time",
			"Last backup time in seconds since epoch (1970)",
			labels, nil,
		),
		BackupInfo: prometheus.NewDesc(
			"sakuracloud_auto_backup_archive_info",
			"A metric with a constant '1' value labeled by backuped archive information",
			backupLabels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *AutoBackupCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Info
	ch <- c.BackupCount
	ch <- c.LastBackupTime
	ch <- c.BackupInfo
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *AutoBackupCollector) Collect(ch chan<- prometheus.Metric) {
	autoBackups, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("auto_backup").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", "can't list autoBackups",
			"err", err,
		)
	}

	var wg sync.WaitGroup

	for i := range autoBackups {
		func(autoBackup *sacloud.AutoBackup) {
			ch <- prometheus.MustNewConstMetric(
				c.Info,
				prometheus.GaugeValue,
				float64(1.0),
				c.autoBackupInfoLabels(autoBackup)...,
			)

			now := time.Now()
			wg.Add(1)
			go func() {
				defer wg.Done()
				c.collectBackupMetrics(ch, autoBackup, now)
			}()
		}(autoBackups[i])
	}

	wg.Wait()
}

func (c *AutoBackupCollector) autoBackupLabels(autoBackup *sacloud.AutoBackup) []string {
	return []string{
		autoBackup.ID.String(),
		autoBackup.Name,
		autoBackup.DiskID.String(),
	}
}

func (c *AutoBackupCollector) autoBackupInfoLabels(autoBackup *sacloud.AutoBackup) []string {
	labels := c.autoBackupLabels(autoBackup)

	return append(labels,
		fmt.Sprintf("%d", autoBackup.MaximumNumberOfArchives),
		flattenBackupSpanWeekdays(autoBackup.BackupSpanWeekdays),
		flattenStringSlice(autoBackup.Tags),
		autoBackup.Description,
	)
}

func (c *AutoBackupCollector) archiveInfoLabels(autoBackup *sacloud.AutoBackup, archive *sacloud.Archive) []string {
	labels := c.autoBackupLabels(autoBackup)
	return append(labels,
		archive.ID.String(),
		archive.Name,
		flattenStringSlice(archive.Tags),
		archive.Description,
	)
}

func (c *AutoBackupCollector) collectBackupMetrics(ch chan<- prometheus.Metric, autoBackup *sacloud.AutoBackup, now time.Time) {
	archives, err := c.client.ListBackups(c.ctx, autoBackup.ZoneName, autoBackup.ID)
	if err != nil {
		c.errors.WithLabelValues("auto_backup").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", "can't list backed up archives",
			"err", err,
		)
		return
	}

	count := 0
	lastTime := int64(0)

	if len(archives) > 0 {
		count = len(archives)
		// asc by CreatedAt
		sort.Slice(archives, func(i, j int) bool { return archives[i].CreatedAt.Before(archives[j].CreatedAt) })
		lastTime = archives[count-1].CreatedAt.Unix()
	}

	ch <- prometheus.MustNewConstMetric(
		c.BackupCount,
		prometheus.GaugeValue,
		float64(count),
		c.autoBackupLabels(autoBackup)...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.LastBackupTime,
		prometheus.GaugeValue,
		float64(lastTime)*1000, // sec to milli-sec
		c.autoBackupLabels(autoBackup)...,
	)

	for _, archive := range archives {
		ch <- prometheus.MustNewConstMetric(
			c.BackupInfo,
			prometheus.GaugeValue,
			float64(1.0),
			c.archiveInfoLabels(autoBackup, archive)...,
		)
	}
}
