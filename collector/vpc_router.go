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

// VPCRouterCollector collects metrics about all servers.
type VPCRouterCollector struct {
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
func NewVPCRouterCollector(logger log.Logger, errors *prometheus.CounterVec, client iaas.VPCRouterClient) *VPCRouterCollector {
	errors.WithLabelValues("vpc_router").Add(0)

	vpcRouterLabels := []string{"id", "name", "zone"}
	vpcRouterInfoLabels := append(vpcRouterLabels, "plan", "ha", "vrid", "vip", "ipaddress1", "ipaddress2", "nw_mask_len", "internet_connection", "tags", "description")
	nicLabels := append(vpcRouterLabels, "nic_index", "vip", "ipaddress1", "ipaddress2", "nw_mask_len")
	s2sPeerLabels := append(vpcRouterLabels, "peer_address", "peer_index")

	return &VPCRouterCollector{
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
	vpcRouters, err := c.client.Find()
	if err != nil {
		c.errors.WithLabelValues("vpc_router").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list vpc_routers",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(vpcRouters))

	for i := range vpcRouters {
		go func(vpcRouter *iaas.VPCRouter) {
			defer wg.Done()

			vpcRouterLabels := c.vpcRouterLabels(vpcRouter)

			var up float64
			if vpcRouter.IsUp() {
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

			if vpcRouter.IsUp() && vpcRouter.HasInterfaces() {

				wg.Add(1)
				go func() {
					defer wg.Done()
					status, err := c.client.Status(vpcRouter.ZoneName, vpcRouter.ID)
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
						if peer.Status == "UP" {
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

				for i := range vpcRouter.Settings.Router.Interfaces {
					// NIC(Receive/Send)
					wg.Add(1)
					go func(i int) {
						c.collectNICMetrics(ch, vpcRouter, i, now)
						wg.Done()
					}(i)
				}
			}

		}(vpcRouters[i])
	}

	wg.Wait()
}

func (c *VPCRouterCollector) vpcRouterLabels(vpcRouter *iaas.VPCRouter) []string {
	return []string{
		vpcRouter.GetStrID(),
		vpcRouter.Name,
		vpcRouter.ZoneName,
	}
}

var vpcRouterPlanMapping = map[int64]string{
	1: "standard",
	2: "premium",
	3: "highspec",
}

func (c *VPCRouterCollector) vpcRouterInfoLabels(vpcRouter *iaas.VPCRouter) []string {
	labels := c.vpcRouterLabels(vpcRouter)

	isHA := "0"
	if !vpcRouter.IsStandardPlan() {
		isHA = "1"
	}

	internetConn := "0"
	if vpcRouter.HasSetting() && vpcRouter.Settings.Router.InternetConnection != nil &&
		vpcRouter.Settings.Router.InternetConnection.Enabled == "True" {
		internetConn = "1"
	}

	vrid := vpcRouter.VRID()
	strVRID := fmt.Sprintf("%d", vrid)
	if vrid < 0 {
		strVRID = ""
	}

	return append(labels,
		vpcRouterPlanMapping[vpcRouter.GetPlanID()],
		isHA,
		strVRID,
		vpcRouter.VirtualIPAddress(),
		vpcRouter.IPAddress1(),
		vpcRouter.IPAddress2(),
		fmt.Sprintf("%d", vpcRouter.NetworkMaskLen()),
		internetConn,
		flattenStringSlice(vpcRouter.Tags),
		vpcRouter.Description,
	)
}

func (c *VPCRouterCollector) nicLabels(vpcRouter *iaas.VPCRouter, index int) []string {
	if len(vpcRouter.Interfaces) <= index {
		return nil
	}

	labels := c.vpcRouterLabels(vpcRouter)
	return append(labels,
		fmt.Sprintf("%d", index),
		vpcRouter.VirtualIPAddressAt(index),
		vpcRouter.IPAddress1At(index),
		vpcRouter.IPAddress2At(index),
		fmt.Sprintf("%d", vpcRouter.NetworkMaskLenAt(index)),
	)
}

func (c *VPCRouterCollector) collectNICMetrics(ch chan<- prometheus.Metric, vpcRouter *iaas.VPCRouter, index int, now time.Time) {
	values, err := c.client.MonitorNIC(vpcRouter.ZoneName, vpcRouter.ID, index, now)
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

	if values.Receive != nil {
		m := prometheus.MustNewConstMetric(
			c.Receive,
			prometheus.GaugeValue,
			values.Receive.Value*8/1000,
			c.nicLabels(vpcRouter, index)...,
		)
		ch <- prometheus.NewMetricWithTimestamp(values.Receive.Time, m)
	}
	if values.Send != nil {
		m := prometheus.MustNewConstMetric(
			c.Send,
			prometheus.GaugeValue,
			values.Send.Value*8/1000,
			c.nicLabels(vpcRouter, index)...,
		)
		ch <- prometheus.NewMetricWithTimestamp(values.Send.Time, m)
	}
}
