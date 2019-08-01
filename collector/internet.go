package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/sakuracloud_exporter/iaas"
)

// InternetCollector collects metrics about all internets.
type InternetCollector struct {
	ctx    context.Context
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.InternetClient

	Info *prometheus.Desc

	In  *prometheus.Desc
	Out *prometheus.Desc
}

// NewInternetCollector returns a new InternetCollector.
func NewInternetCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client iaas.InternetClient) *InternetCollector {
	errors.WithLabelValues("internet").Add(0)

	labels := []string{"id", "name", "zone", "switch_id"}
	infoLabels := append(labels, "bandwidth", "tags", "description")

	return &InternetCollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,
		Info: prometheus.NewDesc(
			"sakuracloud_internet_info",
			"A metric with a constant '1' value labeled by internet information",
			infoLabels, nil,
		),
		In: prometheus.NewDesc(
			"sakuracloud_internet_receive",
			"NIC's receive bytes(unit: Kbps)",
			labels, nil,
		),
		Out: prometheus.NewDesc(
			"sakuracloud_internet_send",
			"NIC's send bytes(unit: Kbps)",
			labels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *InternetCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Info
	ch <- c.In
	ch <- c.Out
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *InternetCollector) Collect(ch chan<- prometheus.Metric) {
	internets, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("internet").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list internets",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(internets))

	for i := range internets {
		func(internet *iaas.Internet) {
			defer wg.Done()

			ch <- prometheus.MustNewConstMetric(
				c.Info,
				prometheus.GaugeValue,
				float64(1.0),
				c.internetInfoLabels(internet)...,
			)

			now := time.Now()
			wg.Add(1)
			go func() {
				c.collectRouterMetrics(ch, internet, now)
				wg.Done()
			}()

		}(internets[i])
	}

	wg.Wait()
}

func (c *InternetCollector) internetLabels(internet *iaas.Internet) []string {
	return []string{
		internet.ID.String(),
		internet.Name,
		internet.ZoneName,
		internet.Switch.ID.String(),
	}
}

func (c *InternetCollector) internetInfoLabels(internet *iaas.Internet) []string {
	labels := c.internetLabels(internet)

	return append(labels,
		fmt.Sprintf("%d", internet.BandWidthMbps),
		flattenStringSlice(internet.Tags),
		internet.Description,
	)
}
func (c *InternetCollector) collectRouterMetrics(ch chan<- prometheus.Metric, internet *iaas.Internet, now time.Time) {

	values, err := c.client.MonitorTraffic(c.ctx, internet.ZoneName, internet.ID, now)
	if err != nil {
		c.errors.WithLabelValues("internet").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get internet's traffic metrics: InternetID=%d", internet.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	in := values.In
	if in > 0 {
		in = in / 1000
	}
	m := prometheus.MustNewConstMetric(
		c.In,
		prometheus.GaugeValue,
		in,
		c.internetLabels(internet)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	out := values.Out
	if out > 0 {
		out = out / 1000
	}
	m = prometheus.MustNewConstMetric(
		c.Out,
		prometheus.GaugeValue,
		out,
		c.internetLabels(internet)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}
