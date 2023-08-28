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
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/newsfeed"
	"github.com/sacloud/sakuracloud_exporter/platform"
	"github.com/stretchr/testify/require"
)

type dummyVPCRouterClient struct {
	find           []*platform.VPCRouter
	findErr        error
	status         *iaas.VPCRouterStatus
	statusErr      error
	monitor        *iaas.MonitorInterfaceValue
	monitorErr     error
	monitorCPU     *iaas.MonitorCPUTimeValue
	monitorCPUErr  error
	maintenance    *newsfeed.FeedItem
	maintenanceErr error
}

func (d *dummyVPCRouterClient) Find(ctx context.Context) ([]*platform.VPCRouter, error) {
	return d.find, d.findErr
}
func (d *dummyVPCRouterClient) Status(ctx context.Context, zone string, id types.ID) (*iaas.VPCRouterStatus, error) {
	return d.status, d.statusErr
}
func (d *dummyVPCRouterClient) MonitorNIC(ctx context.Context, zone string, id types.ID, index int, end time.Time) (*iaas.MonitorInterfaceValue, error) {
	return d.monitor, d.monitorErr
}

func (d *dummyVPCRouterClient) MonitorCPU(ctx context.Context, zone string, id types.ID, end time.Time) (*iaas.MonitorCPUTimeValue, error) {
	return d.monitorCPU, d.monitorCPUErr
}
func (d *dummyVPCRouterClient) MaintenanceInfo(infoURL string) (*newsfeed.FeedItem, error) {
	return d.maintenance, d.maintenanceErr
}

func TestVPCRouterCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewVPCRouterCollector(context.Background(), testLogger, testErrors, &dummyVPCRouterClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Up,
		c.VPCRouterInfo,
		c.CPUTime,
		c.SessionCount,
		c.DHCPLeaseCount,
		c.L2TPSessionCount,
		c.PPTPSessionCount,
		c.SiteToSitePeerStatus,
		c.Receive,
		c.Send,
		c.SessionAnalysis,
		c.MaintenanceScheduled,
		c.MaintenanceInfo,
		c.MaintenanceStartTime,
		c.MaintenanceEndTime,
	}))
}

func TestVPCRouterCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewVPCRouterCollector(context.Background(), testLogger, testErrors, nil)
	monitorTime := time.Unix(1, 0)

	cases := []struct {
		name           string
		in             platform.VPCRouterClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyVPCRouterClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=WARN msg="can't list vpc routers" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyVPCRouterClient{},
			wantMetrics: nil,
		},
		{
			name: "a VPCRouter with activity monitor",
			in: &dummyVPCRouterClient{
				find: []*platform.VPCRouter{
					{
						ZoneName: "is1a",
						VPCRouter: &iaas.VPCRouter{
							ID:             101,
							Name:           "router",
							Description:    "desc",
							Tags:           types.Tags{"tag1", "tag2"},
							PlanID:         types.VPCRouterPlans.Premium,
							InstanceStatus: types.ServerInstanceStatuses.Up,
							Availability:   types.Availabilities.Available,
							Interfaces: []*iaas.VPCRouterInterface{
								{
									Index: 0,
									ID:    200,
								},
								{
									Index: 1,
									ID:    201,
								},
							},
							Settings: &iaas.VPCRouterSetting{
								VRID:                      1,
								InternetConnectionEnabled: true,
								Interfaces: []*iaas.VPCRouterInterfaceSetting{
									{
										VirtualIPAddress: "192.168.0.1",
										IPAddress:        []string{"192.168.0.11", "192.168.0.12"},
										NetworkMaskLen:   24,
										Index:            0,
									},
									{
										VirtualIPAddress: "192.168.1.1",
										IPAddress:        []string{"192.168.1.11", "192.168.1.12"},
										NetworkMaskLen:   24,
										Index:            1,
									},
								},
							},
						},
					},
				},
				status: &iaas.VPCRouterStatus{
					SessionCount: 100,
					DHCPServerLeases: []*iaas.VPCRouterDHCPServerLease{
						{
							IPAddress:  "172.16.0.1",
							MACAddress: "aa:bb:cc:dd:ee:ff",
						},
					},
					L2TPIPsecServerSessions: []*iaas.VPCRouterL2TPIPsecServerSession{
						{
							User:      "user1",
							IPAddress: "172.16.1.1",
							TimeSec:   10,
						},
					},
					PPTPServerSessions: []*iaas.VPCRouterPPTPServerSession{
						{
							User:      "user2",
							IPAddress: "172.16.2.1",
							TimeSec:   20,
						},
					},
					SiteToSiteIPsecVPNPeers: []*iaas.VPCRouterSiteToSiteIPsecVPNPeer{
						{
							Status: "UP",
							Peer:   "172.16.3.1",
						},
					},
					SessionAnalysis: &iaas.VPCRouterSessionAnalysis{
						SourceAddress: []*iaas.VPCRouterStatisticsValue{
							{Name: "localhost", Count: 4},
						},
					},
				},
				monitor: &iaas.MonitorInterfaceValue{
					Time:    monitorTime,
					Receive: 100,
					Send:    200,
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "router",
						"zone": "is1a",
					}),
				},
				{
					desc: c.VPCRouterInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                  "101",
						"name":                "router",
						"zone":                "is1a",
						"plan":                "premium",
						"ha":                  "1",
						"vrid":                "1",
						"vip":                 "192.168.0.1",
						"ipaddress1":          "192.168.0.11",
						"ipaddress2":          "192.168.0.12",
						"nw_mask_len":         "24",
						"internet_connection": "1",
						"tags":                ",tag1,tag2,",
						"description":         "desc",
					}),
				},
				{
					desc: c.SessionCount,
					metric: createGaugeMetric(100, map[string]string{
						"id":   "101",
						"name": "router",
						"zone": "is1a",
					}),
				},
				{
					desc: c.DHCPLeaseCount,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "router",
						"zone": "is1a",
					}),
				},
				{
					desc: c.L2TPSessionCount,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "router",
						"zone": "is1a",
					}),
				},
				{
					desc: c.PPTPSessionCount,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "router",
						"zone": "is1a",
					}),
				},
				{
					desc: c.SiteToSitePeerStatus,
					metric: createGaugeMetric(1, map[string]string{
						"id":           "101",
						"name":         "router",
						"zone":         "is1a",
						"peer_index":   "0",
						"peer_address": "172.16.3.1",
					}),
				},
				{
					desc: c.SessionAnalysis,
					metric: createGaugeMetric(4, map[string]string{
						"id":    "101",
						"name":  "router",
						"zone":  "is1a",
						"type":  "SourceAddress",
						"label": "localhost",
					}),
				},
				{
					desc: c.Receive,
					metric: createGaugeWithTimestamp(float64(100)*8/1000, map[string]string{
						"id":          "101",
						"name":        "router",
						"zone":        "is1a",
						"nic_index":   "0",
						"vip":         "192.168.0.1",
						"ipaddress1":  "192.168.0.11",
						"ipaddress2":  "192.168.0.12",
						"nw_mask_len": "24",
					}, monitorTime),
				},
				{
					desc: c.Receive,
					metric: createGaugeWithTimestamp(float64(100)*8/1000, map[string]string{
						"id":          "101",
						"name":        "router",
						"zone":        "is1a",
						"nic_index":   "1",
						"vip":         "192.168.1.1",
						"ipaddress1":  "192.168.1.11",
						"ipaddress2":  "192.168.1.12",
						"nw_mask_len": "24",
					}, monitorTime),
				},
				{
					desc: c.Send,
					metric: createGaugeWithTimestamp(float64(200)*8/1000, map[string]string{
						"id":          "101",
						"name":        "router",
						"zone":        "is1a",
						"nic_index":   "0",
						"vip":         "192.168.0.1",
						"ipaddress1":  "192.168.0.11",
						"ipaddress2":  "192.168.0.12",
						"nw_mask_len": "24",
					}, monitorTime),
				},
				{
					desc: c.Send,
					metric: createGaugeWithTimestamp(float64(200)*8/1000, map[string]string{
						"id":          "101",
						"name":        "router",
						"zone":        "is1a",
						"nic_index":   "1",
						"vip":         "192.168.1.1",
						"ipaddress1":  "192.168.1.11",
						"ipaddress2":  "192.168.1.12",
						"nw_mask_len": "24",
					}, monitorTime),
				},
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(0, map[string]string{
						"id":   "101",
						"name": "router",
						"zone": "is1a",
					}),
				},
			},
		},
		{
			name: "APIs return error",
			in: &dummyVPCRouterClient{
				find: []*platform.VPCRouter{
					{
						ZoneName: "is1a",
						VPCRouter: &iaas.VPCRouter{
							ID:             101,
							Name:           "router",
							Description:    "desc",
							Tags:           types.Tags{"tag1", "tag2"},
							PlanID:         types.VPCRouterPlans.Premium,
							InstanceStatus: types.ServerInstanceStatuses.Up,
							Availability:   types.Availabilities.Available,
							Interfaces: []*iaas.VPCRouterInterface{
								{Index: 0, ID: 200},
							},
							Settings: &iaas.VPCRouterSetting{
								VRID:                      1,
								InternetConnectionEnabled: true,
								Interfaces: []*iaas.VPCRouterInterfaceSetting{
									{
										VirtualIPAddress: "192.168.0.1",
										IPAddress:        []string{"192.168.0.11", "192.168.0.12"},
										NetworkMaskLen:   24,
										Index:            0,
									},
								},
							},
						},
					},
				},
				statusErr:  errors.New("dummy1"),
				monitorErr: errors.New("dummy2"),
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "router",
						"zone": "is1a",
					}),
				},
				{
					desc: c.VPCRouterInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                  "101",
						"name":                "router",
						"zone":                "is1a",
						"plan":                "premium",
						"ha":                  "1",
						"vrid":                "1",
						"vip":                 "192.168.0.1",
						"ipaddress1":          "192.168.0.11",
						"ipaddress2":          "192.168.0.12",
						"nw_mask_len":         "24",
						"internet_connection": "1",
						"tags":                ",tag1,tag2,",
						"description":         "desc",
					}),
				},
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(0, map[string]string{
						"id":   "101",
						"name": "router",
						"zone": "is1a",
					}),
				},
			},
			wantLogs: []string{
				`level=WARN msg="can't fetch vpc_router's status" err=dummy1`,
				`level=WARN msg="can't get vpc_router's receive bytes: ID=101, NICIndex=0" err=dummy2`,
			},
			wantErrCounter: 2,
		},
		{
			name: "a VPCRouter with maintenance info",
			in: &dummyVPCRouterClient{
				find: []*platform.VPCRouter{
					{
						ZoneName: "is1a",
						VPCRouter: &iaas.VPCRouter{
							ID:                  101,
							Name:                "router",
							Description:         "desc",
							Tags:                types.Tags{"tag1", "tag2"},
							PlanID:              types.VPCRouterPlans.Premium,
							InstanceStatus:      types.ServerInstanceStatuses.Up,
							InstanceHostInfoURL: "http://example.com/maintenance-info-dummy-url",
							Availability:        types.Availabilities.Available,
							Interfaces: []*iaas.VPCRouterInterface{
								{
									Index: 0,
									ID:    200,
								},
								{
									Index: 1,
									ID:    201,
								},
							},
							Settings: &iaas.VPCRouterSetting{
								VRID:                      1,
								InternetConnectionEnabled: true,
								Interfaces: []*iaas.VPCRouterInterfaceSetting{
									{
										VirtualIPAddress: "192.168.0.1",
										IPAddress:        []string{"192.168.0.11", "192.168.0.12"},
										NetworkMaskLen:   24,
										Index:            0,
									},
									{
										VirtualIPAddress: "192.168.1.1",
										IPAddress:        []string{"192.168.1.11", "192.168.1.12"},
										NetworkMaskLen:   24,
										Index:            1,
									},
								},
							},
						},
					},
				},
				maintenance: &newsfeed.FeedItem{
					StrDate:       "947430000", // 2000-01-10
					Description:   "desc",
					StrEventStart: "946652400", // 2000-01-01
					StrEventEnd:   "949244400", // 2000-01-31
					Title:         "dummy-title",
					URL:           "http://example.com/maintenance",
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "router",
						"zone": "is1a",
					}),
				},
				{
					desc: c.VPCRouterInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                  "101",
						"name":                "router",
						"zone":                "is1a",
						"plan":                "premium",
						"ha":                  "1",
						"vrid":                "1",
						"vip":                 "192.168.0.1",
						"ipaddress1":          "192.168.0.11",
						"ipaddress2":          "192.168.0.12",
						"nw_mask_len":         "24",
						"internet_connection": "1",
						"tags":                ",tag1,tag2,",
						"description":         "desc",
					}),
				},
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "router",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "router",
						"zone":        "is1a",
						"info_url":    "http://example.com/maintenance",
						"info_title":  "dummy-title",
						"description": "desc",
						"start_date":  "946652400",
						"end_date":    "949244400",
					}),
				},
				{
					desc: c.MaintenanceStartTime,
					metric: createGaugeMetric(946652400, map[string]string{
						"id":   "101",
						"name": "router",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceEndTime,
					metric: createGaugeMetric(949244400, map[string]string{
						"id":   "101",
						"name": "router",
						"zone": "is1a",
					}),
				},
			},
		},
	}

	for _, tc := range cases {
		initLoggerAndErrors()
		c.logger = testLogger
		c.errors = testErrors
		c.client = tc.in

		collected, err := collectMetrics(c, "vpc_router")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
