// Copyright 2019-2025 The sakuracloud_exporter Authors
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
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/packages-go/newsfeed"
	"github.com/sacloud/sakuracloud_exporter/platform"
)

// MobileGatewayCollector collects metrics about all servers.
type MobileGatewayCollector struct {
	ctx    context.Context
	logger *slog.Logger
	errors *prometheus.CounterVec
	client platform.MobileGatewayClient

	Up                *prometheus.Desc
	MobileGatewayInfo *prometheus.Desc
	Receive           *prometheus.Desc
	Send              *prometheus.Desc

	TrafficControlInfo *prometheus.Desc

	TrafficUplink   *prometheus.Desc
	TrafficDownlink *prometheus.Desc
	TrafficShaping  *prometheus.Desc

	MaintenanceScheduled *prometheus.Desc
	MaintenanceInfo      *prometheus.Desc
	MaintenanceStartTime *prometheus.Desc
	MaintenanceEndTime   *prometheus.Desc
}

// NewMobileGatewayCollector returns a new MobileGatewayCollector.
func NewMobileGatewayCollector(ctx context.Context, logger *slog.Logger, errors *prometheus.CounterVec, client platform.MobileGatewayClient) *MobileGatewayCollector {
	errors.WithLabelValues("mobile_gateway").Add(0)

	mobileGatewayLabels := []string{"id", "name", "zone"}
	mobileGatewayInfoLabels := append(mobileGatewayLabels, "internet_connection", "inter_device_communication", "tags", "description")
	nicLabels := append(mobileGatewayLabels, "nic_index", "ipaddress", "nw_mask_len")
	trafficControlInfoLabel := append(mobileGatewayLabels, "traffic_quota_in_mb", "bandwidth_limit_in_kbps", "enable_email", "enable_slack", "slack_url", "auto_traffic_shaping")

	return &MobileGatewayCollector{
		ctx:    ctx,
		logger: logger,
		errors: errors,
		client: client,
		Up: prometheus.NewDesc(
			"sakuracloud_mobile_gateway_up",
			"If 1 the mobile_gateway is up and running, 0 otherwise",
			mobileGatewayLabels, nil,
		),
		MobileGatewayInfo: prometheus.NewDesc(
			"sakuracloud_mobile_gateway_info",
			"A metric with a constant '1' value labeled by mobile_gateway information",
			mobileGatewayInfoLabels, nil,
		),
		Receive: prometheus.NewDesc(
			"sakuracloud_mobile_gateway_nic_receive",
			"MobileGateway's receive bytes(unit: Kbps)",
			nicLabels, nil,
		),
		Send: prometheus.NewDesc(
			"sakuracloud_mobile_gateway_nic_send",
			"MobileGateway's send bytes(unit: Kbps)",
			nicLabels, nil,
		),
		TrafficControlInfo: prometheus.NewDesc(
			"sakuracloud_mobile_gateway_traffic_control_info",
			"A metric with a constant '1' value labeled by traffic-control information",
			trafficControlInfoLabel, nil,
		),
		TrafficUplink: prometheus.NewDesc(
			"sakuracloud_mobile_gateway_traffic_uplink",
			"MobileGateway's uplink bytes(unit: KB)",
			mobileGatewayLabels, nil,
		),
		TrafficDownlink: prometheus.NewDesc(
			"sakuracloud_mobile_gateway_traffic_downlink",
			"MobileGateway's downlink bytes(unit: KB)",
			mobileGatewayLabels, nil,
		),
		TrafficShaping: prometheus.NewDesc(
			"sakuracloud_mobile_gateway_traffic_shaping",
			"If 1 the traffic is shaped, 0 otherwise",
			mobileGatewayLabels, nil,
		),
		MaintenanceScheduled: prometheus.NewDesc(
			"sakuracloud_mobile_gateway_maintenance_scheduled",
			"If 1 the mobile gateway has scheduled maintenance info, 0 otherwise",
			mobileGatewayLabels, nil,
		),
		MaintenanceInfo: prometheus.NewDesc(
			"sakuracloud_mobile_gateway_maintenance_info",
			"A metric with a constant '1' value labeled by maintenance information",
			append(mobileGatewayLabels, "info_url", "info_title", "description", "start_date", "end_date"), nil,
		),
		MaintenanceStartTime: prometheus.NewDesc(
			"sakuracloud_mobile_gateway_maintenance_start",
			"Scheduled maintenance start time in seconds since epoch (1970)",
			mobileGatewayLabels, nil,
		),
		MaintenanceEndTime: prometheus.NewDesc(
			"sakuracloud_mobile_gateway_maintenance_end",
			"Scheduled maintenance end time in seconds since epoch (1970)",
			mobileGatewayLabels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *MobileGatewayCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.MobileGatewayInfo
	ch <- c.Receive
	ch <- c.Send
	ch <- c.TrafficControlInfo
	ch <- c.TrafficUplink
	ch <- c.TrafficDownlink
	ch <- c.TrafficShaping

	ch <- c.MaintenanceScheduled
	ch <- c.MaintenanceInfo
	ch <- c.MaintenanceStartTime
	ch <- c.MaintenanceEndTime
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *MobileGatewayCollector) Collect(ch chan<- prometheus.Metric) {
	mobileGateways, err := c.client.Find(c.ctx)
	if err != nil {
		c.errors.WithLabelValues("mobile_gateway").Add(1)
		c.logger.Warn(
			"can't list mobile_gateways",
			slog.Any("err", err),
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(mobileGateways))

	for i := range mobileGateways {
		func(mobileGateway *platform.MobileGateway) {
			defer wg.Done()

			mobileGatewayLabels := c.mobileGatewayLabels(mobileGateway)

			var up float64
			if mobileGateway.InstanceStatus.IsUp() {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Up,
				prometheus.GaugeValue,
				up,
				mobileGatewayLabels...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.MobileGatewayInfo,
				prometheus.GaugeValue,
				float64(1.0),
				c.mobileGatewayInfoLabels(mobileGateway)...,
			)
			if mobileGateway.Availability.IsAvailable() && mobileGateway.InstanceStatus.IsUp() {
				// TrafficControlInfo
				wg.Add(1)
				go func() {
					c.collectTrafficControlInfo(ch, mobileGateway)
					wg.Done()
				}()

				// TrafficStatus
				wg.Add(1)
				go func() {
					c.collectTrafficStatus(ch, mobileGateway)
					wg.Done()
				}()

				// collect metrics
				now := time.Now()

				for i := range mobileGateway.Interfaces {
					// NIC(Receive/Send)
					wg.Add(1)
					go func(i int) {
						c.collectNICMetrics(ch, mobileGateway, i, now)
						wg.Done()
					}(i)
				}

				// maintenance info
				var maintenanceScheduled float64
				if mobileGateway.InstanceHostInfoURL != "" {
					maintenanceScheduled = 1.0
					wg.Add(1)
					go func() {
						c.collectMaintenanceInfo(ch, mobileGateway)
						wg.Done()
					}()
				}
				ch <- prometheus.MustNewConstMetric(
					c.MaintenanceScheduled,
					prometheus.GaugeValue,
					maintenanceScheduled,
					mobileGatewayLabels...,
				)
			}
		}(mobileGateways[i])
	}

	wg.Wait()
}

func (c *MobileGatewayCollector) mobileGatewayLabels(mobileGateway *platform.MobileGateway) []string {
	return []string{
		mobileGateway.ID.String(),
		mobileGateway.Name,
		mobileGateway.ZoneName,
	}
}

func (c *MobileGatewayCollector) mobileGatewayInfoLabels(mobileGateway *platform.MobileGateway) []string {
	labels := c.mobileGatewayLabels(mobileGateway)

	internetConnection := "0"
	if mobileGateway.InternetConnectionEnabled {
		internetConnection = "1"
	}

	interDeviceCommunication := "0"
	if mobileGateway.InterDeviceCommunicationEnabled {
		interDeviceCommunication = "1"
	}

	return append(labels,
		internetConnection,
		interDeviceCommunication,
		flattenStringSlice(mobileGateway.Tags),
		mobileGateway.Description,
	)
}

func (c *MobileGatewayCollector) nicLabels(mobileGateway *platform.MobileGateway, index int) []string {
	if len(mobileGateway.Interfaces) <= index {
		return nil
	}

	maskLen := mobileGateway.Interfaces[index].SubnetNetworkMaskLen
	strMaskLen := ""
	if maskLen > 0 {
		strMaskLen = fmt.Sprintf("%d", maskLen)
	}

	labels := c.mobileGatewayLabels(mobileGateway)
	return append(labels,
		fmt.Sprintf("%d", index),
		mobileGateway.Interfaces[index].IPAddress,
		strMaskLen,
	)
}

func (c *MobileGatewayCollector) collectTrafficControlInfo(ch chan<- prometheus.Metric, mobileGateway *platform.MobileGateway) {
	info, err := c.client.TrafficControl(c.ctx, mobileGateway.ZoneName, mobileGateway.ID)
	if err != nil {
		c.errors.WithLabelValues("mobile_gateway").Add(1)
		c.logger.Warn(
			fmt.Sprintf("can't get mobile_gateway's traffic control config: ID=%d", mobileGateway.ID),
			slog.Any("err", err),
		)
		return
	}
	if info == nil {
		return
	}

	enableEmail := "0"
	if info.EmailNotifyEnabled {
		enableEmail = "1"
	}

	enableSlack := "0"
	slackURL := ""
	if info.SlackNotifyEnabled {
		enableSlack = "1"
		slackURL = info.SlackNotifyWebhooksURL
	}

	autoTrafficShaping := "0"
	if info.AutoTrafficShaping {
		autoTrafficShaping = "1"
	}

	labels := append(c.mobileGatewayLabels(mobileGateway),
		fmt.Sprintf("%d", info.TrafficQuotaInMB),
		fmt.Sprintf("%d", info.BandWidthLimitInKbps),
		enableEmail,
		enableSlack,
		slackURL,
		autoTrafficShaping,
	)

	ch <- prometheus.MustNewConstMetric(
		c.TrafficControlInfo,
		prometheus.GaugeValue,
		float64(1.0),
		labels...,
	)
}

func (c *MobileGatewayCollector) collectTrafficStatus(ch chan<- prometheus.Metric, mobileGateway *platform.MobileGateway) {
	status, err := c.client.TrafficStatus(c.ctx, mobileGateway.ZoneName, mobileGateway.ID)
	if err != nil {
		c.errors.WithLabelValues("mobile_gateway").Add(1)
		c.logger.Warn(
			fmt.Sprintf("can't get mobile_gateway's traffic status: ID=%d", mobileGateway.ID),
			slog.Any("err", err),
		)
		return
	}
	if status == nil {
		return
	}

	labels := c.mobileGatewayLabels(mobileGateway)

	trafficShaping := 0
	if status.TrafficShaping {
		trafficShaping = 1
	}
	ch <- prometheus.MustNewConstMetric(
		c.TrafficUplink,
		prometheus.GaugeValue,
		float64(status.UplinkBytes),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.TrafficDownlink,
		prometheus.GaugeValue,
		float64(status.DownlinkBytes),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		c.TrafficShaping,
		prometheus.GaugeValue,
		float64(trafficShaping),
		labels...,
	)
}

func (c *MobileGatewayCollector) collectNICMetrics(ch chan<- prometheus.Metric, mobileGateway *platform.MobileGateway, index int, now time.Time) {
	values, err := c.client.MonitorNIC(c.ctx, mobileGateway.ZoneName, mobileGateway.ID, index, now)
	if err != nil {
		c.errors.WithLabelValues("mobile_gateway").Add(1)
		c.logger.Warn(
			fmt.Sprintf("can't get mobile_gateway's receive bytes: ID=%d, NICIndex=%d", mobileGateway.ID, index),
			slog.Any("err", err),
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
		c.nicLabels(mobileGateway, index)...,
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
		c.nicLabels(mobileGateway, index)...,
	)
	ch <- prometheus.NewMetricWithTimestamp(values.Time, m)
}

func (c *MobileGatewayCollector) maintenanceInfoLabels(resource *platform.MobileGateway, info *newsfeed.FeedItem) []string {
	labels := c.mobileGatewayLabels(resource)

	return append(labels,
		info.URL,
		info.Title,
		info.Description,
		fmt.Sprintf("%d", info.EventStart().Unix()),
		fmt.Sprintf("%d", info.EventEnd().Unix()),
	)
}

func (c *MobileGatewayCollector) collectMaintenanceInfo(ch chan<- prometheus.Metric, resource *platform.MobileGateway) {
	if resource.InstanceHostInfoURL == "" {
		return
	}
	info, err := c.client.MaintenanceInfo(resource.InstanceHostInfoURL)
	if err != nil {
		c.errors.WithLabelValues("mobile_gateway").Add(1)
		c.logger.Warn(
			fmt.Sprintf("can't get mobile gateway's maintenance info: ID=%d", resource.ID),
			slog.Any("err", err),
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
		c.mobileGatewayLabels(resource)...,
	)
	// end
	ch <- prometheus.MustNewConstMetric(
		c.MaintenanceEndTime,
		prometheus.GaugeValue,
		float64(info.EventEnd().Unix()),
		c.mobileGatewayLabels(resource)...,
	)
}
