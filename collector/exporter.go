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
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// ExporterCollector collects metrics, mostly runtime, about this exporter in general.
type ExporterCollector struct {
	ctx       context.Context
	logger    log.Logger
	version   string
	revision  string
	goVersion string
	startTime time.Time

	StartTime *prometheus.Desc
	BuildInfo *prometheus.Desc
}

//logger, Version, Revision, BuildDate, GoVersion, StartTime

// NewExporterCollector returns a new ExporterCollector.
func NewExporterCollector(ctx context.Context, logger log.Logger, version string, revision string, goVersion string, startTime time.Time) *ExporterCollector {
	return &ExporterCollector{
		ctx:    ctx,
		logger: logger,

		version:   version,
		revision:  revision,
		goVersion: goVersion,
		startTime: startTime,

		StartTime: prometheus.NewDesc(
			"sakuracloud_exporter_start_time",
			"Unix timestamp of the start time",
			nil, nil,
		),
		BuildInfo: prometheus.NewDesc(
			"sakuracloud_exporter_build_info",
			"A metric with a constant '1' value labeled by version, revision, and branch from which the node_exporter was built.",
			[]string{"verison", "revision", "goversion"}, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *ExporterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.StartTime
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *ExporterCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log(
		"starttime", c.startTime.Unix(),
		"version", c.version,
		"revision", c.revision,
		"goVersion", c.goVersion,
		"startTime", c.startTime,
	)

	ch <- prometheus.MustNewConstMetric(
		c.StartTime,
		prometheus.GaugeValue,
		float64(c.startTime.Unix()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.BuildInfo,
		prometheus.GaugeValue,
		1.0,
		c.version, c.revision, c.goVersion,
	)
}
