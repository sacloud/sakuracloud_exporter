// Copyright 2019-2020 The sakuracloud_exporter Authors
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
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/sakuracloud_exporter/iaas"
)

// LocalRouterCollector collects metrics about all localRouters.
type LocalRouterCollector struct {
	ctx    context.Context
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.LocalRouterClient

	Up              *prometheus.Desc
	LocalRouterInfo *prometheus.Desc
	SwitchInfo      *prometheus.Desc
	NetworkInfo     *prometheus.Desc
	PeerInfo        *prometheus.Desc
	PeerUp          *prometheus.Desc
	StaticRouteInfo *prometheus.Desc

	ReceiveBytesPerSec *prometheus.Desc
	SendBytesPerSec    *prometheus.Desc
}

// NewLocalRouterCollector returns a new LocalRouterCollector.
func NewLocalRouterCollector(ctx context.Context, logger log.Logger, errors *prometheus.CounterVec, client iaas.LocalRouterClient) *LocalRouterCollector {
	errors.WithLabelValues("local_router").Add(0)

	localRouterLabels := []string{"id", "name"}
	localRouterInfoLabels := append(localRouterLabels, "tags", "description")
	localRouterSwitchInfoLabels := append(localRouterLabels, "category", "code", "zone_id")
	localRouterServerNetworkInfoLabels := append(localRouterLabels, "vip", "ipaddress1", "ipaddress2", "nw_mask_len", "vrid")
	localRouterPeerLabels := append(localRouterLabels, "peer_index", "peer_id")
	localRouterPeerInfoLabels := append(localRouterPeerLabels, "enabled", "description")
	localRouterStaticRouteInfoLabels := append(localRouterLabels, "route_index", "prefix", "next_hop")

	return &LocalRouterCollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,
		Up: prometheus.NewDesc(
			"sakuracloud_local_router_up",
			"If 1 the LocalRouter is available, 0 otherwise",
			localRouterLabels, nil,
		),
		LocalRouterInfo: prometheus.NewDesc(
			"sakuracloud_local_router_info",
			"A metric with a constant '1' value labeled by localRouter information",
			localRouterInfoLabels, nil,
		),
		SwitchInfo: prometheus.NewDesc(
			"sakuracloud_local_router_switch_info",
			"A metric with a constant '1' value labeled by localRouter connected switch information",
			localRouterSwitchInfoLabels, nil,
		),
		NetworkInfo: prometheus.NewDesc(
			"sakuracloud_local_router_network_info",
			"A metric with a constant '1' value labeled by network information of the localRouter",
			localRouterServerNetworkInfoLabels, nil,
		),
		PeerInfo: prometheus.NewDesc(
			"sakuracloud_local_router_peer_info",
			"A metric with a constant '1' value labeled by peer information",
			localRouterPeerInfoLabels, nil,
		),
		PeerUp: prometheus.NewDesc(
			"sakuracloud_local_router_peer_up",
			"If 1 the Peer is available, 0 otherwise",
			localRouterPeerLabels, nil,
		),
		StaticRouteInfo: prometheus.NewDesc(
			"sakuracloud_local_router_static_route_info",
			"A metric with a constant '1' value labeled by static route information",
			localRouterStaticRouteInfoLabels, nil,
		),
		ReceiveBytesPerSec: prometheus.NewDesc(
			"sakuracloud_local_router_receive_per_sec",
			"Receive bytes per seconds",
			localRouterLabels, nil,
		),
		SendBytesPerSec: prometheus.NewDesc(
			"sakuracloud_local_router_send_per_sec",
			"Send bytes per seconds",
			localRouterLabels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *LocalRouterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.LocalRouterInfo
	ch <- c.SwitchInfo
	ch <- c.NetworkInfo
	ch <- c.PeerInfo
	ch <- c.PeerUp
	ch <- c.StaticRouteInfo
	ch <- c.ReceiveBytesPerSec
	ch <- c.SendBytesPerSec
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *LocalRouterCollector) Collect(ch chan<- prometheus.Metric) {
	localRouters, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("local_router").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list localRouters",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(localRouters))

	for i := range localRouters {
		func(localRouter *sacloud.LocalRouter) {
			defer wg.Done()

			localRouterLabels := c.localRouterLabels(localRouter)

			var up float64
			if localRouter.Availability.IsAvailable() {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Up,
				prometheus.GaugeValue,
				up,
				localRouterLabels...,
			)

			c.collectLocalRouterInfo(ch, localRouter)
			if localRouter.Switch != nil {
				c.collectSwitchInfo(ch, localRouter)
			}
			if localRouter.Interface != nil {
				c.collectNetworkInfo(ch, localRouter)
			}

			wg.Add(1)
			go func() {
				c.collectPeerInfo(ch, localRouter)
				wg.Done()
			}()

			for i := range localRouter.StaticRoutes {
				c.collectStaticRouteInfo(ch, localRouter, i)
			}

			if localRouter.Availability.IsAvailable() {
				now := time.Now()

				wg.Add(1)
				go func() {
					c.collectLocalRouterMetrics(ch, localRouter, now)
					wg.Done()
				}()
			}

		}(localRouters[i])
	}

	wg.Wait()
}

func (c *LocalRouterCollector) localRouterLabels(localRouter *sacloud.LocalRouter) []string {
	return []string{
		localRouter.ID.String(),
		localRouter.Name,
	}
}

func (c *LocalRouterCollector) collectLocalRouterInfo(ch chan<- prometheus.Metric, localRouter *sacloud.LocalRouter) {
	labels := append(c.localRouterLabels(localRouter),
		flattenStringSlice(localRouter.Tags),
		localRouter.Description,
	)

	ch <- prometheus.MustNewConstMetric(
		c.LocalRouterInfo,
		prometheus.GaugeValue,
		float64(1.0),
		labels...,
	)
}

func (c *LocalRouterCollector) collectSwitchInfo(ch chan<- prometheus.Metric, localRouter *sacloud.LocalRouter) {
	labels := append(c.localRouterLabels(localRouter),
		localRouter.Switch.Category,
		localRouter.Switch.Code,
		localRouter.Switch.ZoneID,
	)

	ch <- prometheus.MustNewConstMetric(
		c.SwitchInfo,
		prometheus.GaugeValue,
		float64(1.0),
		labels...,
	)
}

func (c *LocalRouterCollector) collectNetworkInfo(ch chan<- prometheus.Metric, localRouter *sacloud.LocalRouter) {
	labels := append(c.localRouterLabels(localRouter),
		localRouter.Interface.VirtualIPAddress,
		localRouter.Interface.IPAddress[0],
		localRouter.Interface.IPAddress[1],
		fmt.Sprintf("%d", localRouter.Interface.NetworkMaskLen),
		fmt.Sprintf("%d", localRouter.Interface.VRID),
	)

	ch <- prometheus.MustNewConstMetric(
		c.NetworkInfo,
		prometheus.GaugeValue,
		float64(1.0),
		labels...,
	)
}

func (c *LocalRouterCollector) collectPeerInfo(ch chan<- prometheus.Metric, localRouter *sacloud.LocalRouter) {

	//localRouterPeerLabels := append(localRouterLabels, "peer_index", "peer_id")
	//localRouterPeerInfoLabels := append(localRouterPeerLabels, "enabled", "description")

	healthStatus, err := c.client.Health(c.ctx, localRouter.ID)
	if err != nil {
		c.errors.WithLabelValues("local_router").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't read health status of the localRouter[%s]", localRouter.ID.String()),
			"err", err,
		)
		return
	}

	for i, peer := range localRouter.Peers {
		peerStatus := c.getPeerStatus(healthStatus.Peers, peer.ID)
		if peerStatus != nil {

			labels := append(c.localRouterLabels(localRouter),
				fmt.Sprintf("%d", i),
				peer.ID.String(),
			)

			up := float64(1.0)
			if strings.ToLower(string(peerStatus.Status)) != "up" {
				up = 0
			}
			ch <- prometheus.MustNewConstMetric(
				c.PeerUp,
				prometheus.GaugeValue,
				up,
				labels...,
			)

			enabled := "0"
			if peer.Enabled {
				enabled = "1"
			}
			infoLabels := append(labels, enabled, peer.Description)
			ch <- prometheus.MustNewConstMetric(
				c.PeerInfo,
				prometheus.GaugeValue,
				float64(1.0),
				infoLabels...,
			)
		}
	}
}

func (c *LocalRouterCollector) getPeerStatus(peerStatuses []*sacloud.LocalRouterHealthPeer, peerID types.ID) *sacloud.LocalRouterHealthPeer {
	for _, peer := range peerStatuses {
		if peer.ID == peerID {
			return peer
		}
	}
	return nil
}

func (c *LocalRouterCollector) collectStaticRouteInfo(ch chan<- prometheus.Metric, localRouter *sacloud.LocalRouter, staticRouteIndex int) {
	labels := append(c.localRouterLabels(localRouter),
		fmt.Sprintf("%d", staticRouteIndex),
		localRouter.StaticRoutes[staticRouteIndex].Prefix,
		localRouter.StaticRoutes[staticRouteIndex].NextHop,
	)

	ch <- prometheus.MustNewConstMetric(
		c.StaticRouteInfo,
		prometheus.GaugeValue,
		float64(1.0),
		labels...,
	)
}

func (c *LocalRouterCollector) collectLocalRouterMetrics(ch chan<- prometheus.Metric, localRouter *sacloud.LocalRouter, now time.Time) {

	values, err := c.client.Monitor(c.ctx, localRouter.ID, now)
	if err != nil {
		c.errors.WithLabelValues("local_router").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get localRouter's metrics: LocalRouterID=%d", localRouter.ID),
			"err", err,
		)
		return
	}
	if values == nil {
		return
	}

	m := prometheus.MustNewConstMetric(
		c.ReceiveBytesPerSec,
		prometheus.GaugeValue,
		values.ReceiveBytesPerSec*8, // byte per sec -> bps(bit)
		c.localRouterLabels(localRouter)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)

	m = prometheus.MustNewConstMetric(
		c.SendBytesPerSec,
		prometheus.GaugeValue,
		values.SendBytesPerSec*8, // byte per sec -> bps(bit)
		c.localRouterLabels(localRouter)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}
