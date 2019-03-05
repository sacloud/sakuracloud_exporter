package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/sacloud/sakuracloud_exporter/iaas"
)

// SIMCollector collects metrics about all sims.
type SIMCollector struct {
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.SIMClient

	Up      *prometheus.Desc
	SIMInfo *prometheus.Desc

	CurrentMonthTraffic *prometheus.Desc
	Uplink              *prometheus.Desc
	Downlink            *prometheus.Desc
}

// NewSIMCollector returns a new SIMCollector.
func NewSIMCollector(logger log.Logger, errors *prometheus.CounterVec, client iaas.SIMClient) *SIMCollector {
	errors.WithLabelValues("sim").Add(0)

	simLabels := []string{"id", "name"}
	simInfoLabels := append(simLabels, "imei_lock",
		"registerd_date", "activated_date", "deactivated_date",
		"ipaddress", "simgroup_id", "carriers", "tags", "description")

	return &SIMCollector{
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
		CurrentMonthTraffic: prometheus.NewDesc(
			"sakuracloud_sim_current_month_traffic",
			"Current month traffic (unit: Kbps)",
			simLabels, nil,
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

	ch <- c.CurrentMonthTraffic
	ch <- c.Uplink
	ch <- c.Downlink
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *SIMCollector) Collect(ch chan<- prometheus.Metric) {
	sims, err := c.client.Find()
	if err != nil {
		c.errors.WithLabelValues("sim").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list sims",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(sims))

	for i := range sims {
		go func(sim *sacloud.SIM) {
			defer wg.Done()

			simLabels := c.simLabels(sim)

			var up float64
			if sim.Status.SIMInfo.SessionStatus == "UP" {
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

			if sim.Status.SIMInfo.SessionStatus == "UP" {
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

func (c *SIMCollector) simLabels(sim *sacloud.SIM) []string {
	return []string{
		sim.GetStrID(),
		sim.Name,
	}
}

func (c *SIMCollector) collectSIMInfo(ch chan<- prometheus.Metric, sim *sacloud.SIM) {
	simConfig, err := c.client.GetNetworkOperatorConfig(sim.ID)
	if err != nil {
		c.errors.WithLabelValues("sim").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get sim's network operator config: SIMID=%d", sim.ID),
			"err", err,
		)
		return
	}
	var carriers []string
	for _, config := range simConfig.NetworkOperatorConfigs {
		if config.Allow {
			carriers = append(carriers, config.Name)
		}
	}

	simInfo := sim.Status.SIMInfo

	imeiLock := "0"
	if simInfo.IMEILock {
		imeiLock = "1"
	}

	var registerdDate, activatedDate, deactivatedDate int64
	if simInfo.RegisteredDate != nil {
		registerdDate = simInfo.RegisteredDate.Unix() * 1000
	}
	if simInfo.ActivatedDate != nil {
		activatedDate = simInfo.ActivatedDate.Unix() * 1000
	}
	if simInfo.DeactivatedDate != nil {
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

func (c *SIMCollector) collectSIMMetrics(ch chan<- prometheus.Metric, sim *sacloud.SIM, now time.Time) {

	values, err := c.client.MonitorTraffic(sim.ID, now)
	if err != nil {
		c.errors.WithLabelValues("sim").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get sim's metrics: SIMID=%d", sim.ID),
			"err", err,
		)
		return
	}
	if len(values) == 0 {
		return
	}

	for _, v := range values {
		if v.Uplink != nil && v.Uplink.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.Uplink,
				prometheus.GaugeValue,
				v.Uplink.Value/1000,
				c.simLabels(sim)...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.Uplink.Time, m)
		}
		if v.Downlink != nil && v.Downlink.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.Downlink,
				prometheus.GaugeValue,
				v.Downlink.Value/1000,
				c.simLabels(sim)...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.Downlink.Time, m)
		}
	}
}
