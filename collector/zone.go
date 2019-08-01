package collector

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/sakuracloud_exporter/iaas"
)

// ZoneCollector collects metrics about the account.
type ZoneCollector struct {
	ctx    context.Context
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.ZoneClient

	ZoneInfo *prometheus.Desc
}

// NewZoneCollector returns a new ZoneCollector.
func NewZoneCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client iaas.ZoneClient) *ZoneCollector {
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
		level.Warn(c.logger).Log(
			"msg", "can't get zone info",
			"err", err,
		)
		return
	}

	for _, zone := range zones {
		labels := []string{
			zone.ID.String(),
			zone.Name,
			zone.Description,
			zone.Region.ID.String(),
			zone.Region.Name,
		}

		ch <- prometheus.MustNewConstMetric(
			c.ZoneInfo,
			prometheus.GaugeValue,
			1.0,
			labels...,
		)

	}
}
