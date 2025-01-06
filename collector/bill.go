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
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/sakuracloud_exporter/platform"
)

// BillCollector collects metrics about the account.
type BillCollector struct {
	ctx    context.Context
	logger *slog.Logger
	errors *prometheus.CounterVec
	client platform.BillClient

	Amount *prometheus.Desc
}

// NewBillCollector returns a new BillCollector.
func NewBillCollector(ctx context.Context, logger *slog.Logger, errors *prometheus.CounterVec, client platform.BillClient) *BillCollector {
	errors.WithLabelValues("bill").Add(0)

	labels := []string{"member_id"}

	return &BillCollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,

		Amount: prometheus.NewDesc(
			"sakuracloud_bill_amount",
			"Amount billed for the month",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *BillCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Amount
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *BillCollector) Collect(ch chan<- prometheus.Metric) {
	bill, err := c.client.Read(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("bill").Add(1)
		c.logger.Warn(
			"can't get bill",
			slog.Any("err", err),
		)
		return
	}

	if bill != nil {
		labels := []string{bill.MemberID}

		// Amount
		ch <- prometheus.MustNewConstMetric(
			c.Amount,
			prometheus.GaugeValue,
			float64(bill.Amount),
			labels...,
		)
	}
}
