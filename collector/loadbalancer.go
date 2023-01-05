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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/newsfeed"
	"github.com/sacloud/sakuracloud_exporter/platform"
)

// LoadBalancerCollector collects metrics about all servers.
type LoadBalancerCollector struct {
	ctx    context.Context
	logger log.Logger
	errors *prometheus.CounterVec
	client platform.LoadBalancerClient

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

	MaintenanceScheduled *prometheus.Desc
	MaintenanceInfo      *prometheus.Desc
	MaintenanceStartTime *prometheus.Desc
	MaintenanceEndTime   *prometheus.Desc
}

// NewLoadBalancerCollector returns a new LoadBalancerCollector.
func NewLoadBalancerCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client platform.LoadBalancerClient) *LoadBalancerCollector {
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
		MaintenanceScheduled: prometheus.NewDesc(
			"sakuracloud_loadbalancer_maintenance_scheduled",
			"If 1 the loadbalancer has scheduled maintenance info, 0 otherwise",
			lbLabels, nil,
		),
		MaintenanceInfo: prometheus.NewDesc(
			"sakuracloud_loadbalancer_maintenance_info",
			"A metric with a constant '1' value labeled by maintenance information",
			append(lbLabels, "info_url", "info_title", "description", "start_date", "end_date"), nil,
		),
		MaintenanceStartTime: prometheus.NewDesc(
			"sakuracloud_loadbalancer_maintenance_start",
			"Scheduled maintenance start time in seconds since epoch (1970)",
			lbLabels, nil,
		),
		MaintenanceEndTime: prometheus.NewDesc(
			"sakuracloud_loadbalancer_maintenance_end",
			"Scheduled maintenance end time in seconds since epoch (1970)",
			lbLabels, nil,
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

	ch <- c.MaintenanceScheduled
	ch <- c.MaintenanceInfo
	ch <- c.MaintenanceStartTime
	ch <- c.MaintenanceEndTime
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *LoadBalancerCollector) Collect(ch chan<- prometheus.Metric) {
	lbs, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("loadbalancer").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", "can't list loadbalancers",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(lbs))

	for i := range lbs {
		func(lb *platform.LoadBalancer) {
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

			if lb.Availability.IsAvailable() && lb.InstanceStatus.IsUp() {
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

				// maintenance info
				var maintenanceScheduled float64
				if lb.InstanceHostInfoURL != "" {
					maintenanceScheduled = 1.0
					wg.Add(1)
					go func() {
						c.collectMaintenanceInfo(ch, lb)
						wg.Done()
					}()
				}
				ch <- prometheus.MustNewConstMetric(
					c.MaintenanceScheduled,
					prometheus.GaugeValue,
					maintenanceScheduled,
					lbLabels...,
				)
			}
		}(lbs[i])
	}

	wg.Wait()
}

func (c *LoadBalancerCollector) lbLabels(lb *platform.LoadBalancer) []string {
	return []string{
		lb.ID.String(),
		lb.Name,
		lb.ZoneName,
	}
}

var loadBalancerPlanMapping = map[types.ID]string{
	types.LoadBalancerPlans.Standard: "standard",
	types.LoadBalancerPlans.HighSpec: "highspec",
}

func (c *LoadBalancerCollector) lbInfoLabels(lb *platform.LoadBalancer) []string {
	labels := c.lbLabels(lb)

	isHA := "0"
	if lb.PlanID == types.LoadBalancerPlans.HighSpec {
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

func (c *LoadBalancerCollector) vipLabels(lb *platform.LoadBalancer, index int) []string {
	if len(lb.VirtualIPAddresses) <= index {
		return nil
	}
	labels := c.lbLabels(lb)
	return append(labels,
		fmt.Sprintf("%d", index),
		lb.VirtualIPAddresses[index].VirtualIPAddress,
	)
}

func (c *LoadBalancerCollector) vipInfoLabels(lb *platform.LoadBalancer, index int) []string {
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

func (c *LoadBalancerCollector) serverLabels(lb *platform.LoadBalancer, vipIndex int, serverIndex int) []string {
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

func (c *LoadBalancerCollector) serverInfoLabels(lb *platform.LoadBalancer, vipIndex int, serverIndex int) []string {
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

func (c *LoadBalancerCollector) collectNICMetrics(ch chan<- prometheus.Metric, lb *platform.LoadBalancer, now time.Time) {
	values, err := c.client.MonitorNIC(c.ctx, lb.ZoneName, lb.ID, now)
	if err != nil {
		c.errors.WithLabelValues("loadbalancer").Add(1)
		level.Warn(c.logger).Log( //nolint
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

func getVIPStatus(status []*iaas.LoadBalancerStatus, vip string) *iaas.LoadBalancerStatus {
	for _, s := range status {
		if s.VirtualIPAddress == vip {
			return s
		}
	}
	return nil
}

func getServerStatus(status []*iaas.LoadBalancerServerStatus, ip string) *iaas.LoadBalancerServerStatus {
	for _, s := range status {
		if s.IPAddress == ip {
			return s
		}
	}
	return nil
}

func (c *LoadBalancerCollector) collectLBStatus(ch chan<- prometheus.Metric, lb *platform.LoadBalancer) {
	status, err := c.client.Status(c.ctx, lb.ZoneName, lb.ID)
	if err != nil {
		c.errors.WithLabelValues("loadbalancer").Add(1)
		level.Warn(c.logger).Log( //nolint
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
		if vipStatus == nil {
			continue
		}
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

func (c *LoadBalancerCollector) maintenanceInfoLabels(resource *platform.LoadBalancer, info *newsfeed.FeedItem) []string {
	labels := c.lbLabels(resource)

	return append(labels,
		info.URL,
		info.Title,
		info.Description,
		fmt.Sprintf("%d", info.EventStart().Unix()),
		fmt.Sprintf("%d", info.EventEnd().Unix()),
	)
}

func (c *LoadBalancerCollector) collectMaintenanceInfo(ch chan<- prometheus.Metric, resource *platform.LoadBalancer) {
	if resource.InstanceHostInfoURL == "" {
		return
	}
	info, err := c.client.MaintenanceInfo(resource.InstanceHostInfoURL)
	if err != nil {
		c.errors.WithLabelValues("loadbalancer").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", fmt.Sprintf("can't get lb's maintenance info: ID=%d", resource.ID),
			"err", err,
		)
		return
	}

	infoLabels := c.maintenanceInfoLabels(resource, info)

	// info
	ch <- prometheus.MustNewConstMetric(
		c.MaintenanceInfo,
		prometheus.GaugeValue,
		1.0,
		infoLabels...,
	)
	// start
	ch <- prometheus.MustNewConstMetric(
		c.MaintenanceStartTime,
		prometheus.GaugeValue,
		float64(info.EventStart().Unix()),
		c.lbLabels(resource)...,
	)
	// end
	ch <- prometheus.MustNewConstMetric(
		c.MaintenanceEndTime,
		prometheus.GaugeValue,
		float64(info.EventEnd().Unix()),
		c.lbLabels(resource)...,
	)
}
