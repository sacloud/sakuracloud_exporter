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

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/sakuracloud_exporter/platform"
)

// ZoneCollector collects metrics about the account.
type ZoneCollector struct {
	ctx    context.Context
	logger log.Logger
	errors *prometheus.CounterVec
	client platform.ZoneClient

	ZoneInfo *prometheus.Desc
}

// NewZoneCollector returns a new ZoneCollector.
func NewZoneCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client platform.ZoneClient) *ZoneCollector {
	errors.WithLabelValues("zone").Add(0)

	labels := []string{"id", "name", "description", "region_id", "region_name"}

	return &ZoneCollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,
		ZoneInfo: prometheus.NewDesc(
			"sakuracloud_zone_info",
			"A metric with a constant '1' value labeled by id, name, description, region_id and region_name",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *ZoneCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.ZoneInfo
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *ZoneCollector) Collect(ch chan<- prometheus.Metric) {
	zones, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("zone").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", "can't get zone info",
			"err", err,
		)
		return
	}

	for _, zone := range zones {
		var regionID, regionName string
		if zone.Region != nil {
			regionID = zone.Region.ID.String()
			regionName = zone.Region.Name
		}
		labels := []string{
			zone.ID.String(),
			zone.Name,
			zone.Description,
			regionID,
			regionName,
		}

		ch <- prometheus.MustNewConstMetric(
			c.ZoneInfo,
			prometheus.GaugeValue,
			1.0,
			labels...,
		)
	}
}
