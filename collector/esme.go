// Copyright 2019-2025 The sakuracloud_exporter Authors
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
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/sakuracloud_exporter/platform"
)

// ESMECollector collects metrics about all esme.
type ESMECollector struct {
	ctx    context.Context
	logger *slog.Logger
	errors *prometheus.CounterVec
	client platform.ESMEClient

	ESMEInfo     *prometheus.Desc
	MessageCount *prometheus.Desc
}

// NewESMECollector returns a new ESMECollector.
func NewESMECollector(ctx context.Context, logger *slog.Logger, errors *prometheus.CounterVec, client platform.ESMEClient) *ESMECollector {
	errors.WithLabelValues("esme").Add(0)

	labels := []string{"id", "name"}
	infoLabels := append(labels, "tags", "description")
	messageLabels := append(labels, "status")

	return &ESMECollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,
		ESMEInfo: prometheus.NewDesc(
			"sakuracloud_esme_info",
			"A metric with a constant '1' value labeled by ESME information",
			infoLabels, nil,
		),
		MessageCount: prometheus.NewDesc(
			"sakuracloud_esme_message_count",
			"A count of messages handled by ESME",
			messageLabels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *ESMECollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.ESMEInfo
	ch <- c.MessageCount
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *ESMECollector) Collect(ch chan<- prometheus.Metric) {
	searched, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("esme").Add(1)
		c.logger.Warn(
			"can't list ESME",
			slog.Any("err", err),
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(searched))

	for i := range searched {
		func(esme *iaas.ESME) {
			defer wg.Done()

			c.collectESMEInfo(ch, esme)

			wg.Add(1)
			go func() {
				c.collectLogs(ch, esme)
				wg.Done()
			}()
		}(searched[i])
	}

	wg.Wait()
}

func (c *ESMECollector) esmeLabels(esme *iaas.ESME) []string {
	return []string{
		esme.ID.String(),
		esme.Name,
	}
}

func (c *ESMECollector) collectESMEInfo(ch chan<- prometheus.Metric, esme *iaas.ESME) {
	labels := append(c.esmeLabels(esme),
		flattenStringSlice(esme.Tags),
		esme.Description,
	)

	ch <- prometheus.MustNewConstMetric(
		c.ESMEInfo,
		prometheus.GaugeValue,
		float64(1.0),
		labels...,
	)
}

func (c *ESMECollector) collectLogs(ch chan<- prometheus.Metric, esme *iaas.ESME) {
	logs, err := c.client.Logs(c.ctx, esme.ID)
	if err != nil {
		c.errors.WithLabelValues("esme").Add(1)
		c.logger.Warn(
			fmt.Sprintf("can't collect logs of the esme[%s]", esme.ID.String()),
			slog.Any("err", err),
		)
		return
	}

	labels := c.esmeLabels(esme)
	labelsForAll := append(labels, "All")

	ch <- prometheus.MustNewConstMetric(
		c.MessageCount,
		prometheus.GaugeValue,
		float64(len(logs)),
		labelsForAll...,
	)

	// count logs per status
	statusCounts := make(map[string]int)
	for _, l := range logs {
		if _, ok := statusCounts[l.Status]; !ok {
			statusCounts[l.Status] = 0
		}
		statusCounts[l.Status]++
	}

	for key, v := range statusCounts {
		labelsPerStatus := append(labels, key)
		ch <- prometheus.MustNewConstMetric(
			c.MessageCount,
			prometheus.GaugeValue,
			float64(v),
			labelsPerStatus...,
		)
	}
}
