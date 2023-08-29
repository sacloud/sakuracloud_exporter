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
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/sakuracloud_exporter/platform"
)

// SIMCollector collects metrics about all sims.
type SIMCollector struct {
	ctx    context.Context
	logger *slog.Logger
	errors *prometheus.CounterVec
	client platform.SIMClient

	Up      *prometheus.Desc
	SIMInfo *prometheus.Desc

	Uplink   *prometheus.Desc
	Downlink *prometheus.Desc
}

// NewSIMCollector returns a new SIMCollector.
func NewSIMCollector(ctx context.Context, logger *slog.Logger, errors *prometheus.CounterVec, client platform.SIMClient) *SIMCollector {
	errors.WithLabelValues("sim").Add(0)

	simLabels := []string{"id", "name"}
	simInfoLabels := append(simLabels, "imei_lock",
		"registered_date", "activated_date", "deactivated_date",
		"ipaddress", "simgroup_id", "carriers", "tags", "description")

	return &SIMCollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,
		Up: prometheus.NewDesc(
			"sakuracloud_sim_session_up",
			"If 1 the session is up and running, 0 otherwise",
			simLabels, nil,
		),
		SIMInfo: prometheus.NewDesc(
			"sakuracloud_sim_info",
			"A metric with a constant '1' value labeled by sim information",
			simInfoLabels, nil,
		),
		Uplink: prometheus.NewDesc(
			"sakuracloud_sim_uplink",
			"Uplink traffic (unit: Kbps)",
			simLabels, nil,
		),
		Downlink: prometheus.NewDesc(
			"sakuracloud_sim_downlink",
			"Downlink traffic (unit: Kbps)",
			simLabels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *SIMCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.SIMInfo

	ch <- c.Uplink
	ch <- c.Downlink
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *SIMCollector) Collect(ch chan<- prometheus.Metric) {
	sims, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("sim").Add(1)
		c.logger.Warn(
			"can't list sims",
			slog.Any("err", err),
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(sims))

	for i := range sims {
		func(sim *iaas.SIM) {
			defer wg.Done()

			simLabels := c.simLabels(sim)

			var up float64
			if strings.ToLower(sim.Info.SessionStatus) == "up" {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Up,
				prometheus.GaugeValue,
				up,
				simLabels...,
			)

			wg.Add(1)
			go func() {
				c.collectSIMInfo(ch, sim)
				wg.Done()
			}()

			if sim.Info.SessionStatus == "UP" {
				now := time.Now()

				wg.Add(1)
				go func() {
					c.collectSIMMetrics(ch, sim, now)
					wg.Done()
				}()
			}
		}(sims[i])
	}

	wg.Wait()
}

func (c *SIMCollector) simLabels(sim *iaas.SIM) []string {
	return []string{
		sim.ID.String(),
		sim.Name,
	}
}

func (c *SIMCollector) collectSIMInfo(ch chan<- prometheus.Metric, sim *iaas.SIM) {
	simConfigs, err := c.client.GetNetworkOperatorConfig(c.ctx, sim.ID)
	if err != nil {
		c.errors.WithLabelValues("sim").Add(1)
		c.logger.Warn(
			fmt.Sprintf("can't get sim's network operator config: SIMID=%d", sim.ID),
			slog.Any("err", err),
		)
		return
	}
	var carriers []string
	for _, config := range simConfigs {
		if config.Allow {
			carriers = append(carriers, config.Name)
		}
	}

	simInfo := sim.Info

	imeiLock := "0"
	if simInfo.IMEILock {
		imeiLock = "1"
	}

	var registerdDate, activatedDate, deactivatedDate int64
	if !simInfo.RegisteredDate.IsZero() {
		registerdDate = simInfo.RegisteredDate.Unix() * 1000
	}
	if !simInfo.ActivatedDate.IsZero() {
		activatedDate = simInfo.ActivatedDate.Unix() * 1000
	}
	if !simInfo.DeactivatedDate.IsZero() {
		deactivatedDate = simInfo.DeactivatedDate.Unix() * 1000
	}

	labels := append(c.simLabels(sim),
		imeiLock,
		fmt.Sprintf("%d", registerdDate),
		fmt.Sprintf("%d", activatedDate),
		fmt.Sprintf("%d", deactivatedDate),
		simInfo.IP,
		simInfo.SIMGroupID,
		flattenStringSlice(carriers),
		flattenStringSlice(sim.Tags),
		sim.Description,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SIMInfo,
		prometheus.GaugeValue,
		float64(1.0),
		labels...,
	)
}

func (c *SIMCollector) collectSIMMetrics(ch chan<- prometheus.Metric, sim *iaas.SIM, now time.Time) {
	values, err := c.client.MonitorTraffic(c.ctx, sim.ID, now)
	if err != nil {
		c.errors.WithLabelValues("sim").Add(1)
		c.logger.Warn(
			fmt.Sprintf("can't get sim's metrics: SIMID=%d", sim.ID),
			slog.Any("err", err),
		)
		return
	}
	if values == nil {
		return
	}

	uplink := values.UplinkBPS
	if uplink > 0 {
		uplink /= 1000
	}
	m := prometheus.MustNewConstMetric(
		c.Uplink,
		prometheus.GaugeValue,
		uplink,
		c.simLabels(sim)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	downlink := values.DownlinkBPS
	if downlink > 0 {
		downlink /= 1000
	}
	m = prometheus.MustNewConstMetric(
		c.Downlink,
		prometheus.GaugeValue,
		downlink,
		c.simLabels(sim)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}
