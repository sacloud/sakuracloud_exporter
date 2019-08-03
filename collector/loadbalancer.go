package collector

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/sacloud/sakuracloud_exporter/iaas"
)

// LoadBalancerCollector collects metrics about all servers.
type LoadBalancerCollector struct {
	ctx    context.Context
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
func NewLoadBalancerCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client iaas.LoadBalancerClient) *LoadBalancerCollector {
	errors.WithLabelValues("loadbalancer").Add(0)

	lbLabels := []string{"id", "name", "zone"}
	lbInfoLabels := append(lbLabels, "plan", "ha", "vrid", "ipaddress1", "ipaddress2", "gateway", "nw_mask_len", "tags", "description")
	vipLabels := append(lbLabels, "vip_index", "vip")
	vipInfoLabels := append(vipLabels, "port", "interval", "sorry_server", "description")
	serverLabels := append(vipLabels, "server_index", "ipaddress")
	serverInfoLabels := append(serverLabels, "monitor", "path", "response_code")

	return &LoadBalancerCollector{
		ctx:    ctx,
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
	lbs, err := c.client.Find(c.ctx)
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
		func(lb *iaas.LoadBalancer) {
			defer wg.Done()

			lbLabels := c.lbLabels(lb)

			var up float64
			if lb.InstanceStatus.IsUp() {
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
			for vipIndex := range lb.VirtualIPAddresses {
				ch <- prometheus.MustNewConstMetric(
					c.VIPInfo,
					prometheus.GaugeValue,
					float64(1.0),
					c.vipInfoLabels(lb, vipIndex)...,
				)
			}

			if lb.InstanceStatus.IsUp() {
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
		lb.ID.String(),
		lb.Name,
		lb.ZoneName,
	}
}

var loadBalancerPlanMapping = map[types.ID]string{
	types.LoadBalancerPlans.Standard: "standard",
	types.LoadBalancerPlans.Premium:  "highspec",
}

func (c *LoadBalancerCollector) lbInfoLabels(lb *iaas.LoadBalancer) []string {
	labels := c.lbLabels(lb)

	isHA := "0"
	if lb.PlanID == types.LoadBalancerPlans.Premium {
		isHA = "1"
	}

	ipaddress2 := ""
	if len(lb.IPAddresses) > 1 {
		ipaddress2 = lb.IPAddresses[1]
	}

	return append(labels,
		loadBalancerPlanMapping[lb.GetPlanID()],
		isHA,
		fmt.Sprintf("%d", lb.VRID),
		lb.IPAddresses[0],
		ipaddress2,
		lb.DefaultRoute,
		fmt.Sprintf("%d", lb.NetworkMaskLen),
		flattenStringSlice(lb.Tags),
		lb.Description,
	)
}

func (c *LoadBalancerCollector) vipLabels(lb *iaas.LoadBalancer, index int) []string {
	if len(lb.VirtualIPAddresses) <= index {
		return nil
	}
	labels := c.lbLabels(lb)
	return append(labels,
		fmt.Sprintf("%d", index),
		lb.VirtualIPAddresses[index].VirtualIPAddress,
	)
}

func (c *LoadBalancerCollector) vipInfoLabels(lb *iaas.LoadBalancer, index int) []string {
	if len(lb.VirtualIPAddresses) <= index {
		return nil
	}
	labels := c.vipLabels(lb, index)
	vip := lb.VirtualIPAddresses[index]
	return append(labels,
		vip.Port.String(),
		vip.DelayLoop.String(),
		vip.SorryServer,
		vip.Description,
	)
}

func (c *LoadBalancerCollector) serverLabels(lb *iaas.LoadBalancer, vipIndex int, serverIndex int) []string {
	if len(lb.VirtualIPAddresses) < vipIndex {
		return nil
	}
	vip := lb.VirtualIPAddresses[vipIndex]
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
	if len(lb.VirtualIPAddresses) < vipIndex {
		return nil
	}
	vip := lb.VirtualIPAddresses[vipIndex]
	if len(vip.Servers) < serverIndex {
		return nil
	}
	server := vip.Servers[serverIndex]

	labels := c.serverLabels(lb, vipIndex, serverIndex)
	return append(labels,
		string(server.HealthCheck.Protocol),
		server.HealthCheck.Path,
		server.HealthCheck.ResponseCode.String(),
	)
}

func (c *LoadBalancerCollector) collectNICMetrics(ch chan<- prometheus.Metric, lb *iaas.LoadBalancer, now time.Time) {
	values, err := c.client.MonitorNIC(c.ctx, lb.ZoneName, lb.ID, now)
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

	receive := values.Receive
	if receive > 0 {
		receive = receive * 8 / 1000
	}
	m := prometheus.MustNewConstMetric(
		c.Receive,
		prometheus.GaugeValue,
		receive,
		c.lbLabels(lb)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	send := values.Send
	if send > 0 {
		send = send * 8 / 1000
	}
	m = prometheus.MustNewConstMetric(
		c.Send,
		prometheus.GaugeValue,
		send,
		c.lbLabels(lb)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func getVIPStatus(status []*sacloud.LoadBalancerStatus, vip string) *sacloud.LoadBalancerStatus {
	for _, s := range status {
		if s.VirtualIPAddress == vip {
			return s
		}
	}
	return nil
}

func getServerStatus(status []*sacloud.LoadBalancerServerStatus, ip string) *sacloud.LoadBalancerServerStatus {
	for _, s := range status {
		if s.IPAddress == ip {
			return s
		}
	}
	return nil
}

func (c *LoadBalancerCollector) collectLBStatus(ch chan<- prometheus.Metric, lb *iaas.LoadBalancer) {
	status, err := c.client.Status(c.ctx, lb.ZoneName, lb.ID)
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

	if len(lb.VirtualIPAddresses) == 0 {
		return
	}
	for vipIndex, vip := range lb.VirtualIPAddresses {
		if vip.Servers == nil {
			return
		}
		vipStatus := getVIPStatus(status, vip.VirtualIPAddress)
		ch <- prometheus.MustNewConstMetric(
			c.VIPCPS,
			prometheus.GaugeValue,
			float64(vipStatus.CPS),
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
			serverStatus := getServerStatus(vipStatus.Servers, server.IPAddress)

			up := float64(0.0)
			activeConn := float64(0.0)
			cps := float64(0.0)
			if serverStatus != nil && strings.ToLower(string(serverStatus.Status)) == "up" {
				up = 1.0
				activeConn = float64(serverStatus.ActiveConn)
				cps = float64(serverStatus.CPS)
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
