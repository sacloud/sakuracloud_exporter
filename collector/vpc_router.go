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

// VPCRouterCollector collects metrics about all servers.
type VPCRouterCollector struct {
	ctx    context.Context
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.VPCRouterClient

	Up            *prometheus.Desc
	SessionCount  *prometheus.Desc
	VPCRouterInfo *prometheus.Desc
	Receive       *prometheus.Desc
	Send          *prometheus.Desc

	DHCPLeaseCount       *prometheus.Desc
	L2TPSessionCount     *prometheus.Desc
	PPTPSessionCount     *prometheus.Desc
	SiteToSitePeerStatus *prometheus.Desc
}

// NewVPCRouterCollector returns a new VPCRouterCollector.
func NewVPCRouterCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client iaas.VPCRouterClient) *VPCRouterCollector {
	errors.WithLabelValues("vpc_router").Add(0)

	vpcRouterLabels := []string{"id", "name", "zone"}
	vpcRouterInfoLabels := append(vpcRouterLabels, "plan", "ha", "vrid", "vip", "ipaddress1", "ipaddress2", "nw_mask_len", "internet_connection", "tags", "description")
	nicLabels := append(vpcRouterLabels, "nic_index", "vip", "ipaddress1", "ipaddress2", "nw_mask_len")
	s2sPeerLabels := append(vpcRouterLabels, "peer_address", "peer_index")

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
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *VPCRouterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.VPCRouterInfo
	ch <- c.SessionCount
	ch <- c.DHCPLeaseCount
	ch <- c.L2TPSessionCount
	ch <- c.PPTPSessionCount
	ch <- c.SiteToSitePeerStatus
	ch <- c.Receive
	ch <- c.Send
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *VPCRouterCollector) Collect(ch chan<- prometheus.Metric) {
	vpcRouters, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("vpc_router").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list vpc routers",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(vpcRouters))

	for i := range vpcRouters {
		func(vpcRouter *iaas.VPCRouter) {
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

			if vpcRouter.InstanceStatus.IsUp() && len(vpcRouter.Interfaces) > 0 {

				wg.Add(1)
				go func() {
					defer wg.Done()
					status, err := c.client.Status(c.ctx, vpcRouter.ZoneName, vpcRouter.ID)
					if err != nil {
						c.errors.WithLabelValues("vpc_router").Add(1)
						level.Warn(c.logger).Log(
							"msg", "can't fetch vpc_router's status",
							"err", err,
						)
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
				}()

				// collect metrics
				now := time.Now()

				for _, nic := range vpcRouter.Interfaces {
					// NIC(Receive/Send)
					wg.Add(1)
					go func(nic *sacloud.VPCRouterInterface) {
						c.collectNICMetrics(ch, vpcRouter, nic.Index, now)
						wg.Done()
					}(nic)
				}
			}

		}(vpcRouters[i])
	}

	wg.Wait()
}

func (c *VPCRouterCollector) vpcRouterLabels(vpcRouter *iaas.VPCRouter) []string {
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

func (c *VPCRouterCollector) vpcRouterInfoLabels(vpcRouter *iaas.VPCRouter) []string {
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

func findVPCRouterInterfaceSettingByIndex(settings []*sacloud.VPCRouterInterfaceSetting, index int) *sacloud.VPCRouterInterfaceSetting {
	if settings != nil {
		for _, s := range settings {
			if s.Index == index {
				return s
			}
		}
	}
	return nil
}

func getInterfaceByIndex(interfaces []*sacloud.VPCRouterInterfaceSetting, index int) *sacloud.VPCRouterInterfaceSetting {
	for _, nic := range interfaces {
		if nic.Index == index {
			return nic
		}
	}
	return nil
}

func (c *VPCRouterCollector) nicLabels(vpcRouter *iaas.VPCRouter, index int) []string {
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

func (c *VPCRouterCollector) collectNICMetrics(ch chan<- prometheus.Metric, vpcRouter *iaas.VPCRouter, index int, now time.Time) {
	values, err := c.client.MonitorNIC(c.ctx, vpcRouter.ZoneName, vpcRouter.ID, index, now)
	if err != nil {
		c.errors.WithLabelValues("vpc_router").Add(1)
		level.Warn(c.logger).Log(
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
