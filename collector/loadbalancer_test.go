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
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/newsfeed"
	"github.com/sacloud/sakuracloud_exporter/platform"
	"github.com/stretchr/testify/require"
)

type dummyLoadBalancerClient struct {
	find           []*platform.LoadBalancer
	findErr        error
	status         []*iaas.LoadBalancerStatus
	statusErr      error
	monitor        *iaas.MonitorInterfaceValue
	monitorErr     error
	maintenance    *newsfeed.FeedItem
	maintenanceErr error
}

func (d *dummyLoadBalancerClient) Find(ctx context.Context) ([]*platform.LoadBalancer, error) {
	return d.find, d.findErr
}
func (d *dummyLoadBalancerClient) Status(ctx context.Context, zone string, id types.ID) ([]*iaas.LoadBalancerStatus, error) {
	return d.status, d.statusErr
}
func (d *dummyLoadBalancerClient) MonitorNIC(ctx context.Context, zone string, id types.ID, end time.Time) (*iaas.MonitorInterfaceValue, error) {
	return d.monitor, d.monitorErr
}
func (d *dummyLoadBalancerClient) MaintenanceInfo(infoURL string) (*newsfeed.FeedItem, error) {
	return d.maintenance, d.maintenanceErr
}

func TestLoadBalancerCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewLoadBalancerCollector(context.Background(), testLogger, testErrors, &dummyLoadBalancerClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Up,
		c.LoadBalancerInfo,
		c.Receive,
		c.Send,
		c.VIPInfo,
		c.VIPCPS,
		c.ServerInfo,
		c.ServerUp,
		c.ServerConnection,
		c.ServerCPS,
		c.MaintenanceScheduled,
		c.MaintenanceInfo,
		c.MaintenanceStartTime,
		c.MaintenanceEndTime,
	}))
}

func TestLoadBalancerCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewLoadBalancerCollector(context.Background(), testLogger, testErrors, nil)
	monitorTime := time.Unix(1, 0)

	cases := []struct {
		name           string
		in             platform.LoadBalancerClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyLoadBalancerClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't list loadbalancers" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyLoadBalancerClient{},
			wantMetrics: nil,
		},
		{
			name: "a load balancer",
			in: &dummyLoadBalancerClient{
				find: []*platform.LoadBalancer{
					{
						ZoneName: "is1a",
						LoadBalancer: &iaas.LoadBalancer{
							ID:             101,
							Name:           "loadbalancer",
							Tags:           types.Tags{"tag1", "tag2"},
							Description:    "desc",
							PlanID:         types.LoadBalancerPlans.Standard,
							VRID:           1,
							IPAddresses:    []string{"192.168.0.11"},
							DefaultRoute:   "192.168.0.1",
							NetworkMaskLen: 24,
							Availability:   types.Availabilities.Available,
							InstanceStatus: types.ServerInstanceStatuses.Up,
						},
					},
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
						"zone": "is1a",
					}),
				},
				{
					desc: c.LoadBalancerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "loadbalancer",
						"zone":        "is1a",
						"plan":        "standard",
						"ha":          "0",
						"vrid":        "1",
						"ipaddress1":  "192.168.0.11",
						"ipaddress2":  "",
						"gateway":     "192.168.0.1",
						"nw_mask_len": "24",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(0, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
						"zone": "is1a",
					}),
				},
			},
		},
		{
			name: "a highspec load balancer with activity monitors",
			in: &dummyLoadBalancerClient{
				find: []*platform.LoadBalancer{
					{
						ZoneName: "is1a",
						LoadBalancer: &iaas.LoadBalancer{
							ID:             101,
							Name:           "loadbalancer",
							Tags:           types.Tags{"tag1", "tag2"},
							Description:    "desc",
							PlanID:         types.LoadBalancerPlans.HighSpec,
							VRID:           1,
							IPAddresses:    []string{"192.168.0.11", "192.168.0.12"},
							DefaultRoute:   "192.168.0.1",
							NetworkMaskLen: 24,
							Availability:   types.Availabilities.Available,
							InstanceStatus: types.ServerInstanceStatuses.Up,
							VirtualIPAddresses: []*iaas.LoadBalancerVirtualIPAddress{
								{
									VirtualIPAddress: "192.168.0.101",
									Port:             80,
									DelayLoop:        100,
									SorryServer:      "192.168.0.21",
									Description:      "vip-desc",
									Servers: []*iaas.LoadBalancerServer{
										{
											IPAddress: "192.168.0.201",
											Port:      80,
											Enabled:   true,
											HealthCheck: &iaas.LoadBalancerServerHealthCheck{
												Protocol:     types.LoadBalancerHealthCheckProtocols.HTTP,
												ResponseCode: http.StatusOK,
												Path:         "/index.html",
											},
										},
									},
								},
							},
						},
					},
				},
				status: []*iaas.LoadBalancerStatus{
					{
						VirtualIPAddress: "192.168.0.101",
						Port:             80,
						CPS:              100,
						Servers: []*iaas.LoadBalancerServerStatus{
							{
								IPAddress:  "192.168.0.201",
								Port:       80,
								Status:     types.ServerInstanceStatuses.Up,
								CPS:        200,
								ActiveConn: 300,
							},
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
						"name": "loadbalancer",
						"zone": "is1a",
					}),
				},
				{
					desc: c.LoadBalancerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "loadbalancer",
						"zone":        "is1a",
						"plan":        "highspec",
						"ha":          "1",
						"vrid":        "1",
						"ipaddress1":  "192.168.0.11",
						"ipaddress2":  "192.168.0.12",
						"gateway":     "192.168.0.1",
						"nw_mask_len": "24",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.VIPInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":           "101",
						"name":         "loadbalancer",
						"zone":         "is1a",
						"vip_index":    "0",
						"vip":          "192.168.0.101",
						"port":         "80",
						"interval":     "100",
						"sorry_server": "192.168.0.21",
						"description":  "vip-desc",
					}),
				},
				{
					desc: c.Receive,
					metric: createGaugeWithTimestamp(float64(100)*8/1000, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
						"zone": "is1a",
					}, monitorTime),
				},
				{
					desc: c.Send,
					metric: createGaugeWithTimestamp(float64(200)*8/1000, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
						"zone": "is1a",
					}, monitorTime),
				},
				{
					desc: c.VIPCPS,
					metric: createGaugeMetric(100, map[string]string{
						"id":        "101",
						"name":      "loadbalancer",
						"zone":      "is1a",
						"vip_index": "0",
						"vip":       "192.168.0.101",
					}),
				},
				{
					desc: c.ServerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":            "101",
						"name":          "loadbalancer",
						"zone":          "is1a",
						"vip_index":     "0",
						"vip":           "192.168.0.101",
						"server_index":  "0",
						"ipaddress":     "192.168.0.201",
						"monitor":       "http",
						"path":          "/index.html",
						"response_code": "200",
					}),
				},
				{
					desc: c.ServerUp,
					metric: createGaugeMetric(1, map[string]string{
						"id":           "101",
						"name":         "loadbalancer",
						"zone":         "is1a",
						"vip_index":    "0",
						"vip":          "192.168.0.101",
						"server_index": "0",
						"ipaddress":    "192.168.0.201",
					}),
				},
				{
					desc: c.ServerCPS,
					metric: createGaugeMetric(200, map[string]string{
						"id":           "101",
						"name":         "loadbalancer",
						"zone":         "is1a",
						"vip_index":    "0",
						"vip":          "192.168.0.101",
						"server_index": "0",
						"ipaddress":    "192.168.0.201",
					}),
				},
				{
					desc: c.ServerConnection,
					metric: createGaugeMetric(300, map[string]string{
						"id":           "101",
						"name":         "loadbalancer",
						"zone":         "is1a",
						"vip_index":    "0",
						"vip":          "192.168.0.101",
						"server_index": "0",
						"ipaddress":    "192.168.0.201",
					}),
				},
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(0, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
						"zone": "is1a",
					}),
				},
			},
		},
		{
			name: "status and monitor API return error",
			in: &dummyLoadBalancerClient{
				find: []*platform.LoadBalancer{
					{
						ZoneName: "is1a",
						LoadBalancer: &iaas.LoadBalancer{
							ID:             101,
							Name:           "loadbalancer",
							Tags:           types.Tags{"tag1", "tag2"},
							Description:    "desc",
							PlanID:         types.LoadBalancerPlans.HighSpec,
							VRID:           1,
							IPAddresses:    []string{"192.168.0.11", "192.168.0.12"},
							DefaultRoute:   "192.168.0.1",
							NetworkMaskLen: 24,
							Availability:   types.Availabilities.Available,
							InstanceStatus: types.ServerInstanceStatuses.Up,
							VirtualIPAddresses: []*iaas.LoadBalancerVirtualIPAddress{
								{
									VirtualIPAddress: "192.168.0.101",
									Port:             80,
									DelayLoop:        100,
									SorryServer:      "192.168.0.21",
									Description:      "vip-desc",
									Servers: []*iaas.LoadBalancerServer{
										{
											IPAddress: "192.168.0.201",
											Port:      80,
											Enabled:   true,
											HealthCheck: &iaas.LoadBalancerServerHealthCheck{
												Protocol:     types.LoadBalancerHealthCheckProtocols.HTTP,
												ResponseCode: http.StatusOK,
												Path:         "/index.html",
											},
										},
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
						"name": "loadbalancer",
						"zone": "is1a",
					}),
				},
				{
					desc: c.LoadBalancerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "loadbalancer",
						"zone":        "is1a",
						"plan":        "highspec",
						"ha":          "1",
						"vrid":        "1",
						"ipaddress1":  "192.168.0.11",
						"ipaddress2":  "192.168.0.12",
						"gateway":     "192.168.0.1",
						"nw_mask_len": "24",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.VIPInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":           "101",
						"name":         "loadbalancer",
						"zone":         "is1a",
						"vip_index":    "0",
						"vip":          "192.168.0.101",
						"port":         "80",
						"interval":     "100",
						"sorry_server": "192.168.0.21",
						"description":  "vip-desc",
					}),
				},
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(0, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
						"zone": "is1a",
					}),
				},
			},
			wantLogs: []string{
				`level=warn msg="can't fetch loadbalancer's status: ID: 101" err=dummy1`,
				`level=warn msg="can't get loadbalancer's NIC metrics: ID=101" err=dummy2`,
			},
			wantErrCounter: 2,
		},
		{
			name: "a load balancer with maintenance info",
			in: &dummyLoadBalancerClient{
				find: []*platform.LoadBalancer{
					{
						ZoneName: "is1a",
						LoadBalancer: &iaas.LoadBalancer{
							ID:                  101,
							Name:                "loadbalancer",
							Tags:                types.Tags{"tag1", "tag2"},
							Description:         "desc",
							PlanID:              types.LoadBalancerPlans.Standard,
							VRID:                1,
							IPAddresses:         []string{"192.168.0.11"},
							DefaultRoute:        "192.168.0.1",
							NetworkMaskLen:      24,
							Availability:        types.Availabilities.Available,
							InstanceStatus:      types.ServerInstanceStatuses.Up,
							InstanceHostInfoURL: "http://example.com/maintenance-info-dummy-url",
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
						"name": "loadbalancer",
						"zone": "is1a",
					}),
				},
				{
					desc: c.LoadBalancerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "loadbalancer",
						"zone":        "is1a",
						"plan":        "standard",
						"ha":          "0",
						"vrid":        "1",
						"ipaddress1":  "192.168.0.11",
						"ipaddress2":  "",
						"gateway":     "192.168.0.1",
						"nw_mask_len": "24",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "loadbalancer",
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
						"name": "loadbalancer",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceEndTime,
					metric: createGaugeMetric(949244400, map[string]string{
						"id":   "101",
						"name": "loadbalancer",
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

		collected, err := collectMetrics(c, "loadbalancer")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
