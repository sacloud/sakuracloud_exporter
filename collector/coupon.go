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
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/sakuracloud_exporter/platform"
)

// CouponCollector collects metrics about the account.
type CouponCollector struct {
	ctx    context.Context
	logger log.Logger
	errors *prometheus.CounterVec
	client platform.CouponClient

	Discount      *prometheus.Desc
	RemainingDays *prometheus.Desc
	ExpDate       *prometheus.Desc
	Usable        *prometheus.Desc
}

// NewCouponCollector returns a new CouponCollector.
func NewCouponCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client platform.CouponClient) *CouponCollector {
	errors.WithLabelValues("coupon").Add(0)

	labels := []string{"id", "member_id", "contract_id"}

	return &CouponCollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,

		Discount: prometheus.NewDesc(
			"sakuracloud_coupon_discount",
			"The balance of coupon",
			labels, nil,
		),
		RemainingDays: prometheus.NewDesc(
			"sakuracloud_coupon_remaining_days",
			"The count of coupon's remaining days",
			labels, nil,
		),
		ExpDate: prometheus.NewDesc(
			"sakuracloud_coupon_exp_date",
			"Coupon expiration date in seconds since epoch (1970)",
			labels, nil,
		),
		Usable: prometheus.NewDesc(
			"sakuracloud_coupon_usable",
			"1 if your coupon is usable",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *CouponCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Discount
	ch <- c.RemainingDays
	ch <- c.ExpDate
	ch <- c.Usable
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *CouponCollector) Collect(ch chan<- prometheus.Metric) {
	coupons, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("coupon").Add(1)
		level.Warn(c.logger).Log( // nolint
			"msg", "can't get coupon",
			"err", err,
		)
		return
	}

	for _, coupon := range coupons {
		labels := []string{
			coupon.ID.String(),
			coupon.MemberID,
			fmt.Sprintf("%d", coupon.ContractID),
		}

		now := time.Now()

		// Discount
		ch <- prometheus.MustNewConstMetric(
			c.Discount,
			prometheus.GaugeValue,
			float64(coupon.Discount),
			labels...,
		)

		// RemainingDays
		remainingDays := int(coupon.UntilAt.Sub(now).Hours() / 24)
		if remainingDays < 0 {
			remainingDays = 0
		}
		ch <- prometheus.MustNewConstMetric(
			c.RemainingDays,
			prometheus.GaugeValue,
			float64(remainingDays),
			labels...,
		)

		// Expiration date
		ch <- prometheus.MustNewConstMetric(
			c.ExpDate,
			prometheus.GaugeValue,
			float64(coupon.UntilAt.Unix())*1000,
			labels...,
		)

		// Usable
		var usable float64
		if coupon.Discount > 0 && coupon.AppliedAt.Before(now) && coupon.UntilAt.After(now) {
			usable = 1
		}
		ch <- prometheus.MustNewConstMetric(
			c.Usable,
			prometheus.GaugeValue,
			usable,
			labels...,
		)
	}
}
