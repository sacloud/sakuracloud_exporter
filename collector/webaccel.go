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
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/sakuracloud_exporter/platform"
)

// WebAccelCollector collects metrics about the webaccel's sites.
type WebAccelCollector struct {
	ctx    context.Context
	logger *slog.Logger
	errors *prometheus.CounterVec
	client platform.WebAccelClient

	SiteInfo           *prometheus.Desc
	AccessCount        *prometheus.Desc
	BytesSent          *prometheus.Desc
	CacheMissBytesSent *prometheus.Desc
	CacheHitRatio      *prometheus.Desc
	BytesCacheHitRatio *prometheus.Desc
	Price              *prometheus.Desc

	CertificateExpireDate *prometheus.Desc
}

// NewWebAccelCollector returns a new WebAccelCollector.
func NewWebAccelCollector(ctx context.Context, logger *slog.Logger, errors *prometheus.CounterVec, client platform.WebAccelClient) *WebAccelCollector {
	errors.WithLabelValues("webaccel").Add(0)

	labels := []string{"id"}

	return &WebAccelCollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,
		SiteInfo: prometheus.NewDesc(
			"webaccel_site_info",
			"A metric with a constant '1' value labeled by id, name, domain_type, domain, subdomain",
			[]string{"id", "name", "domain_type", "domain", "subdomain"}, nil,
		),
		AccessCount: prometheus.NewDesc(
			"webaccel_access_count",
			"",
			labels, nil,
		),
		BytesSent: prometheus.NewDesc(
			"webaccel_bytes_sent",
			"",
			labels, nil,
		),
		CacheMissBytesSent: prometheus.NewDesc(
			"webaccel_cache_miss_bytes_sent",
			"",
			labels, nil,
		),
		CacheHitRatio: prometheus.NewDesc(
			"webaccel_cache_hit_ratio",
			"",
			labels, nil,
		),
		BytesCacheHitRatio: prometheus.NewDesc(
			"webaccel_bytes_cache_hit_ratio",
			"",
			labels, nil,
		),
		Price: prometheus.NewDesc(
			"webaccel_price",
			"",
			labels, nil,
		),
		CertificateExpireDate: prometheus.NewDesc(
			"webaccel_cert_expire",
			"Certificate expiration date in seconds since epoch (1970)",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *WebAccelCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.SiteInfo
	ch <- c.AccessCount
	ch <- c.BytesSent
	ch <- c.CacheMissBytesSent
	ch <- c.CacheHitRatio
	ch <- c.BytesCacheHitRatio
	ch <- c.Price
	ch <- c.CertificateExpireDate
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *WebAccelCollector) Collect(ch chan<- prometheus.Metric) {
	sites, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("webaccel").Add(1)
		c.logger.Warn(
			"can't get webAccel info",
			slog.Any("err", err),
		)
		return
	}

	for _, site := range sites {
		labels := []string{
			site.ID,
			site.Name,
			site.DomainType,
			site.Domain,
			site.Subdomain,
		}

		ch <- prometheus.MustNewConstMetric(
			c.SiteInfo,
			prometheus.GaugeValue,
			1.0,
			labels...,
		)

		if site.HasCertificate {
			ch <- prometheus.MustNewConstMetric(
				c.CertificateExpireDate,
				prometheus.GaugeValue,
				float64(site.CertValidNotAfter),
				[]string{site.ID}...,
			)
		}
	}

	usage, err := c.client.Usage(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("webaccel").Add(1)
		c.logger.Warn(
			"can't get webAccel monthly usage",
			slog.Any("err", err),
		)
		return
	}
	for _, u := range usage.MonthlyUsages {
		labels := []string{u.SiteID.String()}

		ch <- prometheus.MustNewConstMetric(
			c.AccessCount,
			prometheus.GaugeValue,
			float64(u.AccessCount),
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesSent,
			prometheus.GaugeValue,
			float64(u.BytesSent),
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CacheMissBytesSent,
			prometheus.GaugeValue,
			float64(u.CacheMissBytesSent),
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.CacheHitRatio,
			prometheus.GaugeValue,
			u.CacheHitRatio,
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesCacheHitRatio,
			prometheus.GaugeValue,
			u.BytesCacheHitRatio,
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			c.Price,
			prometheus.GaugeValue,
			float64(u.Price),
			labels...,
		)
	}
}
