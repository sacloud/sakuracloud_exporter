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

// VPCRouterCollector collects metrics about all servers.
type VPCRouterCollector struct {
	ctx    context.Context
	logger log.Logger
	errors *prometheus.CounterVec
	client platform.VPCRouterClient

	Up            *prometheus.Desc
	SessionCount  *prometheus.Desc
	VPCRouterInfo *prometheus.Desc
	Receive       *prometheus.Desc
	Send          *prometheus.Desc

	CPUTime              *prometheus.Desc
	DHCPLeaseCount       *prometheus.Desc
	L2TPSessionCount     *prometheus.Desc
	PPTPSessionCount     *prometheus.Desc
	SiteToSitePeerStatus *prometheus.Desc

	SessionAnalysis *prometheus.Desc

	MaintenanceScheduled *prometheus.Desc
	MaintenanceInfo      *prometheus.Desc
	MaintenanceStartTime *prometheus.Desc
	MaintenanceEndTime   *prometheus.Desc
}

// NewVPCRouterCollector returns a new VPCRouterCollector.
func NewVPCRouterCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client platform.VPCRouterClient) *VPCRouterCollector {
	errors.WithLabelValues("vpc_router").Add(0)

	vpcRouterLabels := []string{"id", "name", "zone"}
	vpcRouterInfoLabels := append(vpcRouterLabels, "plan", "ha", "vrid", "vip", "ipaddress1", "ipaddress2", "nw_mask_len", "internet_connection", "tags", "description")
	nicLabels := append(vpcRouterLabels, "nic_index", "vip", "ipaddress1", "ipaddress2", "nw_mask_len")
	s2sPeerLabels := append(vpcRouterLabels, "peer_address", "peer_index")
	sessionAnalysisLabels := append(vpcRouterLabels, "type", "label")

	return &VPCRouterCollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,
		Up: prometheus.NewDesc(
			"sakuracloud_vpc_router_up",
			"If 1 the vpc_router is up and running, 0 otherwise",
			vpcRouterLabels, nil,
		),
		SessionCount: prometheus.NewDesc(
			"sakuracloud_vpc_router_session",
			"Current session count",
			vpcRouterLabels, nil,
		),
		VPCRouterInfo: prometheus.NewDesc(
			"sakuracloud_vpc_router_info",
			"A metric with a constant '1' value labeled by vpc_router information",
			vpcRouterInfoLabels, nil,
		),
		CPUTime: prometheus.NewDesc(
			"sakuracloud_vpc_router_cpu_time",
			"VPCRouter's CPU time(unit: ms)",
			vpcRouterLabels, nil,
		),
		DHCPLeaseCount: prometheus.NewDesc(
			"sakuracloud_vpc_router_dhcp_lease",
			"Current DHCPServer lease count",
			vpcRouterLabels, nil,
		),
		L2TPSessionCount: prometheus.NewDesc(
			"sakuracloud_vpc_router_l2tp_session",
			"Current L2TP-IPsec session count",
			vpcRouterLabels, nil,
		),
		PPTPSessionCount: prometheus.NewDesc(
			"sakuracloud_vpc_router_pptp_session",
			"Current PPTP session count",
			vpcRouterLabels, nil,
		),
		SiteToSitePeerStatus: prometheus.NewDesc(
			"sakuracloud_vpc_router_s2s_peer_up",
			"If 1 the vpc_router's site to site peer is up, 0 otherwise",
			s2sPeerLabels, nil,
		),
		Receive: prometheus.NewDesc(
			"sakuracloud_vpc_router_receive",
			"VPCRouter's receive bytes(unit: Kbps)",
			nicLabels, nil,
		),
		Send: prometheus.NewDesc(
			"sakuracloud_vpc_router_send",
			"VPCRouter's receive bytes(unit: Kbps)",
			nicLabels, nil,
		),
		SessionAnalysis: prometheus.NewDesc(
			"sakuracloud_vpc_router_session_analysis",
			"Session statistics for VPC routers",
			sessionAnalysisLabels, nil,
		),
		MaintenanceScheduled: prometheus.NewDesc(
			"sakuracloud_vpc_router_maintenance_scheduled",
			"If 1 the vpc router has scheduled maintenance info, 0 otherwise",
			vpcRouterLabels, nil,
		),
		MaintenanceInfo: prometheus.NewDesc(
			"sakuracloud_vpc_router_maintenance_info",
			"A metric with a constant '1' value labeled by maintenance information",
			append(vpcRouterLabels, "info_url", "info_title", "description", "start_date", "end_date"), nil,
		),
		MaintenanceStartTime: prometheus.NewDesc(
			"sakuracloud_vpc_router_maintenance_start",
			"Scheduled maintenance start time in seconds since epoch (1970)",
			vpcRouterLabels, nil,
		),
		MaintenanceEndTime: prometheus.NewDesc(
			"sakuracloud_vpc_router_maintenance_end",
			"Scheduled maintenance end time in seconds since epoch (1970)",
			vpcRouterLabels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *VPCRouterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.VPCRouterInfo
	ch <- c.CPUTime
	ch <- c.SessionCount
	ch <- c.DHCPLeaseCount
	ch <- c.L2TPSessionCount
	ch <- c.PPTPSessionCount
	ch <- c.SiteToSitePeerStatus
	ch <- c.Receive
	ch <- c.Send
	ch <- c.SessionAnalysis

	ch <- c.MaintenanceScheduled
	ch <- c.MaintenanceInfo
	ch <- c.MaintenanceStartTime
	ch <- c.MaintenanceEndTime
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *VPCRouterCollector) Collect(ch chan<- prometheus.Metric) {
	vpcRouters, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("vpc_router").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", "can't list vpc routers",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(vpcRouters))

	for i := range vpcRouters {
		func(vpcRouter *platform.VPCRouter) {
			defer wg.Done()

			vpcRouterLabels := c.vpcRouterLabels(vpcRouter)

			var up float64
			if vpcRouter.InstanceStatus.IsUp() {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Up,
				prometheus.GaugeValue,
				up,
				vpcRouterLabels...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.VPCRouterInfo,
				prometheus.GaugeValue,
				float64(1.0),
				c.vpcRouterInfoLabels(vpcRouter)...,
			)

			if vpcRouter.Availability.IsAvailable() && vpcRouter.InstanceStatus.IsUp() {
				// collect metrics per resources under server
				now := time.Now()
				// CPU-TIME
				wg.Add(1)
				go func() {
					c.collectCPUTime(ch, vpcRouter, now)
					wg.Done()
				}()

				if len(vpcRouter.Interfaces) > 0 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						status, err := c.client.Status(c.ctx, vpcRouter.ZoneName, vpcRouter.ID)
						if err != nil {
							c.errors.WithLabelValues("vpc_router").Add(1)
							level.Warn(c.logger).Log( //nolint
								"msg", "can't fetch vpc_router's status",
								"err", err,
							)
							return
						}
						if status == nil {
							return
						}

						// Session Count
						ch <- prometheus.MustNewConstMetric(
							c.SessionCount,
							prometheus.GaugeValue,
							float64(status.SessionCount),
							c.vpcRouterLabels(vpcRouter)...,
						)
						// DHCP Server Leases
						ch <- prometheus.MustNewConstMetric(
							c.DHCPLeaseCount,
							prometheus.GaugeValue,
							float64(len(status.DHCPServerLeases)),
							c.vpcRouterLabels(vpcRouter)...,
						)
						// L2TP/IPsec Sessions
						ch <- prometheus.MustNewConstMetric(
							c.L2TPSessionCount,
							prometheus.GaugeValue,
							float64(len(status.L2TPIPsecServerSessions)),
							c.vpcRouterLabels(vpcRouter)...,
						)
						// PPTP Sessions
						ch <- prometheus.MustNewConstMetric(
							c.PPTPSessionCount,
							prometheus.GaugeValue,
							float64(len(status.PPTPServerSessions)),
							c.vpcRouterLabels(vpcRouter)...,
						)
						// Site to Site Peer
						for i, peer := range status.SiteToSiteIPsecVPNPeers {
							up := float64(0)
							if strings.ToLower(peer.Status) == "up" {
								up = float64(1)
							}
							labels := append(c.vpcRouterLabels(vpcRouter),
								peer.Peer,
								fmt.Sprintf("%d", i),
							)

							ch <- prometheus.MustNewConstMetric(
								c.SiteToSitePeerStatus,
								prometheus.GaugeValue,
								up,
								labels...,
							)
						}
						if status.SessionAnalysis != nil {
							sessionAnalysis := map[string][]*iaas.VPCRouterStatisticsValue{
								"SourceAndDestination": status.SessionAnalysis.SourceAndDestination,
								"DestinationAddress":   status.SessionAnalysis.DestinationAddress,
								"DestinationPort":      status.SessionAnalysis.DestinationPort,
								"SourceAddress":        status.SessionAnalysis.SourceAddress,
							}
							for typeName, analysis := range sessionAnalysis {
								for _, v := range analysis {
									labels := append(c.vpcRouterLabels(vpcRouter), typeName, v.Name)
									ch <- prometheus.MustNewConstMetric(
										c.SessionAnalysis,
										prometheus.GaugeValue,
										float64(v.Count),
										labels...,
									)
								}
							}
						}
					}()

					// collect metrics
					for _, nic := range vpcRouter.Interfaces {
						// NIC(Receive/Send)
						wg.Add(1)
						go func(nic *iaas.VPCRouterInterface) {
							c.collectNICMetrics(ch, vpcRouter, nic.Index, now)
							wg.Done()
						}(nic)
					}
				}

				// maintenance info
				var maintenanceScheduled float64
				if vpcRouter.InstanceHostInfoURL != "" {
					maintenanceScheduled = 1.0
					wg.Add(1)
					go func() {
						c.collectMaintenanceInfo(ch, vpcRouter)
						wg.Done()
					}()
				}
				ch <- prometheus.MustNewConstMetric(
					c.MaintenanceScheduled,
					prometheus.GaugeValue,
					maintenanceScheduled,
					vpcRouterLabels...,
				)
			}
		}(vpcRouters[i])
	}

	wg.Wait()
}

func (c *VPCRouterCollector) vpcRouterLabels(vpcRouter *platform.VPCRouter) []string {
	return []string{
		vpcRouter.ID.String(),
		vpcRouter.Name,
		vpcRouter.ZoneName,
	}
}

var vpcRouterPlanMapping = map[types.ID]string{
	types.VPCRouterPlans.Standard: "standard",
	types.VPCRouterPlans.Premium:  "premium",
	types.VPCRouterPlans.HighSpec: "highspec",
}

func (c *VPCRouterCollector) vpcRouterInfoLabels(vpcRouter *platform.VPCRouter) []string {
	labels := c.vpcRouterLabels(vpcRouter)

	isHA := "0"
	if vpcRouter.PlanID != types.VPCRouterPlans.Standard {
		isHA = "1"
	}

	internetConn := "0"
	if vpcRouter.Settings.InternetConnectionEnabled {
		internetConn = "1"
	}

	vrid := vpcRouter.Settings.VRID
	strVRID := fmt.Sprintf("%d", vrid)
	if vrid < 0 {
		strVRID = ""
	}

	var vip, ipaddress1, ipaddress2 string
	var nwMaskLen = "-"
	if nicSetting := findVPCRouterInterfaceSettingByIndex(vpcRouter.Settings.Interfaces, 0); nicSetting != nil {
		vip = nicSetting.VirtualIPAddress
		if len(nicSetting.IPAddress) > 0 {
			ipaddress1 = nicSetting.IPAddress[0]
		}
		if len(nicSetting.IPAddress) > 1 {
			ipaddress2 = nicSetting.IPAddress[1]
		}
		nwMaskLen = fmt.Sprintf("%d", nicSetting.NetworkMaskLen)
	}

	return append(labels,
		vpcRouterPlanMapping[vpcRouter.GetPlanID()],
		isHA,
		strVRID,
		vip,
		ipaddress1,
		ipaddress2,
		nwMaskLen,
		internetConn,
		flattenStringSlice(vpcRouter.Tags),
		vpcRouter.Description,
	)
}

func findVPCRouterInterfaceSettingByIndex(settings []*iaas.VPCRouterInterfaceSetting, index int) *iaas.VPCRouterInterfaceSetting {
	for _, s := range settings {
		if s.Index == index {
			return s
		}
	}
	return nil
}

func getInterfaceByIndex(interfaces []*iaas.VPCRouterInterfaceSetting, index int) *iaas.VPCRouterInterfaceSetting {
	for _, nic := range interfaces {
		if nic.Index == index {
			return nic
		}
	}
	return nil
}

func (c *VPCRouterCollector) nicLabels(vpcRouter *platform.VPCRouter, index int) []string {
	if len(vpcRouter.Interfaces) <= index {
		return nil
	}

	var vip, ipaddress1, ipaddress2 string
	nwMaskLen := ""

	labels := c.vpcRouterLabels(vpcRouter)
	nic := getInterfaceByIndex(vpcRouter.Settings.Interfaces, index)
	if nic != nil {
		vip = nic.VirtualIPAddress
		if len(nic.IPAddress) > 0 {
			ipaddress1 = nic.IPAddress[0]
		}
		if len(nic.IPAddress) > 1 {
			ipaddress2 = nic.IPAddress[1]
		}
		nwMaskLen = fmt.Sprintf("%d", nic.NetworkMaskLen)
	}
	return append(labels,
		fmt.Sprintf("%d", index),
		vip,
		ipaddress1,
		ipaddress2,
		nwMaskLen,
	)
}

func (c *VPCRouterCollector) collectNICMetrics(ch chan<- prometheus.Metric, vpcRouter *platform.VPCRouter, index int, now time.Time) {
	values, err := c.client.MonitorNIC(c.ctx, vpcRouter.ZoneName, vpcRouter.ID, index, now)
	if err != nil {
		c.errors.WithLabelValues("vpc_router").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", fmt.Sprintf("can't get vpc_router's receive bytes: ID=%d, NICIndex=%d", vpcRouter.ID, index),
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
		c.nicLabels(vpcRouter, index)...,
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
		c.nicLabels(vpcRouter, index)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *VPCRouterCollector) collectCPUTime(ch chan<- prometheus.Metric, vpcRouter *platform.VPCRouter, now time.Time) {
	values, err := c.client.MonitorCPU(c.ctx, vpcRouter.ZoneName, vpcRouter.ID, now)
	if err != nil {
		c.errors.WithLabelValues("server").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", fmt.Sprintf("can't get server's CPU-TIME: ID=%d", vpcRouter.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	m := prometheus.MustNewConstMetric(
		c.CPUTime,
		prometheus.GaugeValue,
		values.CPUTime*1000,
		c.vpcRouterLabels(vpcRouter)...,
	)

	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *VPCRouterCollector) maintenanceInfoLabels(resource *platform.VPCRouter, info *newsfeed.FeedItem) []string {
	labels := c.vpcRouterLabels(resource)

	return append(labels,
		info.URL,
		info.Title,
		info.Description,
		fmt.Sprintf("%d", info.EventStart().Unix()),
		fmt.Sprintf("%d", info.EventEnd().Unix()),
	)
}

func (c *VPCRouterCollector) collectMaintenanceInfo(ch chan<- prometheus.Metric, resource *platform.VPCRouter) {
	if resource.InstanceHostInfoURL == "" {
		return
	}
	info, err := c.client.MaintenanceInfo(resource.InstanceHostInfoURL)
	if err != nil {
		c.errors.WithLabelValues("vpc_router").Add(1)
		level.Warn(c.logger).Log( //nolint
			"msg", fmt.Sprintf("can't get vpc router's maintenance info: ID=%d", resource.ID),
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
		c.vpcRouterLabels(resource)...,
	)
	// end
	ch <- prometheus.MustNewConstMetric(
		c.MaintenanceEndTime,
		prometheus.GaugeValue,
		float64(info.EventEnd().Unix()),
		c.vpcRouterLabels(resource)...,
	)
}
