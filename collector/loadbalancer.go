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

// LoadBalancerCollector collects metrics about all servers.
type LoadBalancerCollector struct {
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.LoadBalancerClient

	Up               *prometheus.Desc
	LoadBalancerInfo *prometheus.Desc
	Receive          *prometheus.Desc
	Send             *prometheus.Desc

	VIPInfo *prometheus.Desc
	VIPCPS  *prometheus.Desc

	ServerInfo       *prometheus.Desc
	ServerUp         *prometheus.Desc
	ServerConnection *prometheus.Desc
	ServerCPS        *prometheus.Desc
}

// NewLoadBalancerCollector returns a new LoadBalancerCollector.
func NewLoadBalancerCollector(logger log.Logger, errors *prometheus.CounterVec, client iaas.LoadBalancerClient) *LoadBalancerCollector {
	errors.WithLabelValues("loadbalancer").Add(0)

	lbLabels := []string{"id", "name", "zone"}
	lbInfoLabels := append(lbLabels, "plan", "ha", "vrid", "ipaddress1", "ipaddress2", "gateway", "nw_mask_len", "tags", "description")
	vipLabels := append(lbLabels, "vip_index", "vip")
	vipInfoLabels := append(vipLabels, "port", "interval", "sorry_server", "description")
	serverLabels := append(vipLabels, "server_index", "ipaddress")
	serverInfoLabels := append(serverLabels, "monitor", "path", "response_code")

	return &LoadBalancerCollector{
		logger: logger,
		errors: errors,
		client: client,
		Up: prometheus.NewDesc(
			"sakuracloud_loadbalancer_up",
			"If 1 the loadbalancer is up and running, 0 otherwise",
			lbLabels, nil,
		),
		LoadBalancerInfo: prometheus.NewDesc(
			"sakuracloud_loadbalancer_info",
			"A metric with a constant '1' value labeled by loadbalancer information",
			lbInfoLabels, nil,
		),
		Receive: prometheus.NewDesc(
			"sakuracloud_loadbalancer_receive",
			"Loadbalancer's receive bytes(unit: Kbps)",
			lbLabels, nil,
		),
		Send: prometheus.NewDesc(
			"sakuracloud_loadbalancer_send",
			"Loadbalancer's receive bytes(unit: Kbps)",
			lbLabels, nil,
		),
		VIPInfo: prometheus.NewDesc(
			"sakuracloud_loadbalancer_vip_info",
			"A metric with a constant '1' value labeld by vip information",
			vipInfoLabels, nil,
		),
		VIPCPS: prometheus.NewDesc(
			"sakuracloud_loadbalancer_vip_cps",
			"Connection count per second",
			vipLabels, nil,
		),
		ServerInfo: prometheus.NewDesc(
			"sakuracloud_loadbalancer_server_info",
			"A metric with a constant '1' value labeld by real-server information",
			serverInfoLabels, nil,
		),
		ServerUp: prometheus.NewDesc(
			"sakuracloud_loadbalancer_server_up",
			"If 1 the server is up and running, 0 otherwise",
			serverLabels, nil,
		),
		ServerConnection: prometheus.NewDesc(
			"sakuracloud_loadbalancer_server_connection",
			"Current connection count",
			serverLabels, nil,
		),
		ServerCPS: prometheus.NewDesc(
			"sakuracloud_loadbalancer_server_cps",
			"Connection count per second",
			serverLabels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *LoadBalancerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.LoadBalancerInfo
	ch <- c.Receive
	ch <- c.Send
	ch <- c.VIPInfo
	ch <- c.VIPCPS
	ch <- c.ServerInfo
	ch <- c.ServerUp
	ch <- c.ServerConnection
	ch <- c.ServerCPS
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *LoadBalancerCollector) Collect(ch chan<- prometheus.Metric) {
	lbs, err := c.client.Find()
	if err != nil {
		c.errors.WithLabelValues("loadbalancer").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list loadbalancers",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(lbs))

	for i := range lbs {
		go func(lb *iaas.LoadBalancer) {
			defer wg.Done()

			lbLabels := c.lbLabels(lb)

			var up float64
			if lb.IsUp() {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Up,
				prometheus.GaugeValue,
				up,
				lbLabels...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.LoadBalancerInfo,
				prometheus.GaugeValue,
				float64(1.0),
				c.lbInfoLabels(lb)...,
			)
			if lb.Settings != nil && lb.Settings.LoadBalancer != nil {
				for vipIndex := range lb.Settings.LoadBalancer {
					ch <- prometheus.MustNewConstMetric(
						c.VIPInfo,
						prometheus.GaugeValue,
						float64(1.0),
						c.vipInfoLabels(lb, vipIndex)...,
					)
				}
			}

			if lb.IsUp() {
				now := time.Now()

				// NIC(Receive/Send)
				wg.Add(1)
				go func() {
					c.collectNICMetrics(ch, lb, now)
					wg.Done()
				}()

				// VIP/Server status
				wg.Add(1)
				go func() {
					c.collectLBStatus(ch, lb)
					wg.Done()
				}()

			}

		}(lbs[i])
	}

	wg.Wait()
}

func (c *LoadBalancerCollector) lbLabels(lb *iaas.LoadBalancer) []string {
	return []string{
		lb.GetStrID(),
		lb.Name,
		lb.ZoneName,
	}
}

var loadBalancerPlanMapping = map[int64]string{
	1: "standard",
	2: "highspec",
}

func (c *LoadBalancerCollector) lbInfoLabels(lb *iaas.LoadBalancer) []string {
	labels := c.lbLabels(lb)

	isHA := "0"
	if lb.IsHA() {
		isHA = "1"
	}

	return append(labels,
		loadBalancerPlanMapping[lb.GetPlanID()],
		isHA,
		fmt.Sprintf("%d", lb.Remark.VRRP.VRID),
		lb.IPAddress1(),
		lb.IPAddress2(),
		lb.Remark.Network.DefaultRoute,
		fmt.Sprintf("%d", lb.Remark.Network.NetworkMaskLen),
		flattenStringSlice(lb.Tags),
		lb.Description,
	)
}

func (c *LoadBalancerCollector) vipLabels(lb *iaas.LoadBalancer, index int) []string {
	if len(lb.Settings.LoadBalancer) <= index {
		return nil
	}
	labels := c.lbLabels(lb)
	return append(labels,
		fmt.Sprintf("%d", index),
		lb.Settings.LoadBalancer[index].VirtualIPAddress,
	)
}

func (c *LoadBalancerCollector) vipInfoLabels(lb *iaas.LoadBalancer, index int) []string {
	if len(lb.Settings.LoadBalancer) <= index {
		return nil
	}
	labels := c.vipLabels(lb, index)
	vip := lb.Settings.LoadBalancer[index]
	return append(labels,
		vip.Port,
		vip.DelayLoop,
		vip.SorryServer,
		vip.Description,
	)
}

func (c *LoadBalancerCollector) serverLabels(lb *iaas.LoadBalancer, vipIndex int, serverIndex int) []string {
	if len(lb.Settings.LoadBalancer) < vipIndex {
		return nil
	}
	vip := lb.Settings.LoadBalancer[vipIndex]
	if len(vip.Servers) < serverIndex {
		return nil
	}
	server := vip.Servers[serverIndex]

	labels := c.vipLabels(lb, vipIndex)
	return append(labels,
		fmt.Sprintf("%d", serverIndex),
		server.IPAddress,
	)
}

func (c *LoadBalancerCollector) serverInfoLabels(lb *iaas.LoadBalancer, vipIndex int, serverIndex int) []string {
	if len(lb.Settings.LoadBalancer) < vipIndex {
		return nil
	}
	vip := lb.Settings.LoadBalancer[vipIndex]
	if len(vip.Servers) < serverIndex {
		return nil
	}
	server := vip.Servers[serverIndex]

	labels := c.serverLabels(lb, vipIndex, serverIndex)
	return append(labels,
		server.HealthCheck.Protocol,
		server.HealthCheck.Path,
		server.HealthCheck.Status,
	)
}

func (c *LoadBalancerCollector) collectNICMetrics(ch chan<- prometheus.Metric, lb *iaas.LoadBalancer, now time.Time) {
	values, err := c.client.MonitorNIC(lb.ZoneName, lb.ID, now)
	if err != nil {
		c.errors.WithLabelValues("loadbalancer").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get loadbalancer's NIC metrics: ID=%d", lb.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	if values.Receive != nil {
		m := prometheus.MustNewConstMetric(
			c.Receive,
			prometheus.GaugeValue,
			values.Receive.Value*8/1000,
			c.lbLabels(lb)...,
		)
		ch <- prometheus.NewMetricWithTimestamp(values.Receive.Time, m)
	}
	if values.Send != nil {
		m := prometheus.MustNewConstMetric(
			c.Send,
			prometheus.GaugeValue,
			values.Send.Value*8/1000,
			c.lbLabels(lb)...,
		)
		ch <- prometheus.NewMetricWithTimestamp(values.Send.Time, m)
	}
}

func (c *LoadBalancerCollector) collectLBStatus(ch chan<- prometheus.Metric, lb *iaas.LoadBalancer) {
	status, err := c.client.Status(lb.ZoneName, lb.ID)
	if err != nil {
		c.errors.WithLabelValues("loadbalancer").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't fetch loadbalancer's status: ID: %d", lb.ID),
			"err", err,
		)
		return
	}
	if status == nil {
		return
	}

	if lb.Settings == nil || lb.Settings.LoadBalancer == nil {
		return
	}
	for vipIndex, vip := range lb.Settings.LoadBalancer {
		if vip.Servers == nil {
			return
		}
		vipStatus := status.Get(vip.VirtualIPAddress)
		ch <- prometheus.MustNewConstMetric(
			c.VIPCPS,
			prometheus.GaugeValue,
			float64(vipStatus.NumCPS()),
			c.vipLabels(lb, vipIndex)...,
		)
		for serverIndex, server := range vip.Servers {

			// ServerInfo
			ch <- prometheus.MustNewConstMetric(
				c.ServerInfo,
				prometheus.GaugeValue,
				float64(1.0),
				c.serverInfoLabels(lb, vipIndex, serverIndex)...,
			)

			serverStatus := vipStatus.Get(server.IPAddress)

			up := float64(0.0)
			activeConn := float64(0.0)
			cps := float64(0.0)
			if serverStatus != nil && serverStatus.Status == "UP" {
				up = 1.0
				activeConn = float64(serverStatus.NumActiveConn())
				cps = float64(serverStatus.NumCPS())
			}

			ch <- prometheus.MustNewConstMetric(
				c.ServerUp,
				prometheus.GaugeValue,
				up,
				c.serverLabels(lb, vipIndex, serverIndex)...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.ServerConnection,
				prometheus.GaugeValue,
				activeConn,
				c.serverLabels(lb, vipIndex, serverIndex)...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.ServerCPS,
				prometheus.GaugeValue,
				cps,
				c.serverLabels(lb, vipIndex, serverIndex)...,
			)
		}
	}
}
