package collector

import (
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
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.InternetClient

	Info *prometheus.Desc

	In  *prometheus.Desc
	Out *prometheus.Desc
}

// NewInternetCollector returns a new InternetCollector.
func NewInternetCollector(logger log.Logger, errors *prometheus.CounterVec, client iaas.InternetClient) *InternetCollector {
	errors.WithLabelValues("internet").Add(0)

	labels := []string{"id", "name", "zone", "switch_id"}
	infoLabels := append(labels, "bandwidth", "tags", "description")

	return &InternetCollector{
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
	internets, err := c.client.Find()
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
		go func(internet *iaas.Internet) {
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
		internet.GetStrID(),
		internet.Name,
		internet.ZoneName,
		internet.Switch.GetStrID(),
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

	values, err := c.client.MonitorTraffic(internet.ZoneName, internet.ID, now)
	if err != nil {
		c.errors.WithLabelValues("internet").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get internet's traffic metrics: InternetID=%d", internet.ID),
			"err", err,
		)
		return
	}
	if len(values) == 0 {
		return
	}

	for _, v := range values {
		if v.In != nil && v.In.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.In,
				prometheus.GaugeValue,
				v.In.Value/1000,
				c.internetLabels(internet)...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.In.Time, m)
		}
		if v.Out != nil && v.Out.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.Out,
				prometheus.GaugeValue,
				v.Out.Value/1000,
				c.internetLabels(internet)...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.Out.Time, m)
		}
	}
}
