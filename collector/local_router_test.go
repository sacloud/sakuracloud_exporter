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
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/sacloud/sakuracloud_exporter/iaas"
	"github.com/stretchr/testify/require"
)

type dummyLocalRouterClient struct {
	find       []*sacloud.LocalRouter
	findErr    error
	health     *sacloud.LocalRouterHealth
	healthErr  error
	monitor    *sacloud.MonitorLocalRouterValue
	monitorErr error
}

func (d *dummyLocalRouterClient) Find(ctx context.Context) ([]*sacloud.LocalRouter, error) {
	return d.find, d.findErr
}

func (d *dummyLocalRouterClient) Health(ctx context.Context, id types.ID) (*sacloud.LocalRouterHealth, error) {
	return d.health, d.healthErr
}

func (d *dummyLocalRouterClient) Monitor(ctx context.Context, id types.ID, end time.Time) (*sacloud.MonitorLocalRouterValue, error) {
	return d.monitor, d.monitorErr
}

func TestLocalRouterCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewLocalRouterCollector(context.Background(), testLogger, testErrors, &dummyLocalRouterClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Up,
		c.LocalRouterInfo,
		c.SwitchInfo,
		c.NetworkInfo,
		c.PeerInfo,
		c.PeerUp,
		c.StaticRouteInfo,
		c.ReceiveBytesPerSec,
		c.SendBytesPerSec,
	}))
}

func TestLocalRouterCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewLocalRouterCollector(context.Background(), testLogger, testErrors, nil)
	monitorTime := time.Unix(1, 0)

	cases := []struct {
		name           string
		in             iaas.LocalRouterClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyLocalRouterClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't list localRouters" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyLocalRouterClient{},
			wantMetrics: nil,
		},
		{
			name: "a local router",
			in: &dummyLocalRouterClient{
				find: []*sacloud.LocalRouter{
					{
						ID:           101,
						Name:         "local-router",
						Tags:         types.Tags{"tag1", "tag2"},
						Description:  "desc",
						Availability: types.Availabilities.Available,
						Switch: &sacloud.LocalRouterSwitch{
							Code:     "201",
							Category: "cloud",
							ZoneID:   "is1a",
						},
						Interface: &sacloud.LocalRouterInterface{
							VirtualIPAddress: "192.0.2.1",
							IPAddress:        []string{"192.0.2.11", "192.0.2.12"},
							NetworkMaskLen:   24,
							VRID:             100,
						},
						StaticRoutes: []*sacloud.LocalRouterStaticRoute{
							{
								Prefix:  "10.0.0.0/24",
								NextHop: "192.0.2.101",
							},
							{
								Prefix:  "10.0.1.0/24",
								NextHop: "192.0.2.102",
							},
						},
					},
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "local-router",
					}),
				},
				{
					desc: c.LocalRouterInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "local-router",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.SwitchInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":       "101",
						"name":     "local-router",
						"code":     "201",
						"category": "cloud",
						"zone_id":  "is1a",
					}),
				},
				{
					desc: c.NetworkInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "local-router",
						"vip":         "192.0.2.1",
						"ipaddress1":  "192.0.2.11",
						"ipaddress2":  "192.0.2.12",
						"nw_mask_len": "24",
						"vrid":        "100",
					}),
				},
				{
					desc: c.StaticRouteInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "local-router",
						"route_index": "0",
						"prefix":      "10.0.0.0/24",
						"next_hop":    "192.0.2.101",
					}),
				},
				{
					desc: c.StaticRouteInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "local-router",
						"route_index": "1",
						"prefix":      "10.0.1.0/24",
						"next_hop":    "192.0.2.102",
					}),
				},
			},
		},
		{
			name: "a local router with peers",
			in: &dummyLocalRouterClient{
				find: []*sacloud.LocalRouter{
					{
						ID:           101,
						Name:         "local-router",
						Tags:         types.Tags{"tag1", "tag2"},
						Description:  "desc",
						Availability: types.Availabilities.Available,
						Peers: []*sacloud.LocalRouterPeer{
							{
								ID:          201,
								SecretKey:   "dummy",
								Enabled:     true,
								Description: "desc201",
							},
							{
								ID:          202,
								SecretKey:   "dummy",
								Enabled:     true,
								Description: "desc202",
							},
						},
					},
				},
				health: &sacloud.LocalRouterHealth{
					Peers: []*sacloud.LocalRouterHealthPeer{
						{
							ID:     201,
							Status: "UP",
							Routes: []string{"10.0.0.0/24"},
						},
						{
							ID:     202,
							Status: "UP",
							Routes: []string{"10.0.1.0/24"},
						},
					},
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "local-router",
					}),
				},
				{
					desc: c.LocalRouterInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "local-router",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.PeerUp,
					metric: createGaugeMetric(1, map[string]string{
						"id":         "101",
						"name":       "local-router",
						"peer_index": "0",
						"peer_id":    "201",
					}),
				},
				{
					desc: c.PeerUp,
					metric: createGaugeMetric(1, map[string]string{
						"id":         "101",
						"name":       "local-router",
						"peer_index": "1",
						"peer_id":    "202",
					}),
				},
				{
					desc: c.PeerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "local-router",
						"peer_index":  "0",
						"peer_id":     "201",
						"enabled":     "1",
						"description": "desc201",
					}),
				},
				{
					desc: c.PeerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "local-router",
						"peer_index":  "1",
						"peer_id":     "202",
						"enabled":     "1",
						"description": "desc202",
					}),
				},
			},
		},
		{
			name: "a local router with activities",
			in: &dummyLocalRouterClient{
				find: []*sacloud.LocalRouter{
					{
						ID:           101,
						Name:         "local-router",
						Tags:         types.Tags{"tag1", "tag2"},
						Description:  "desc",
						Availability: types.Availabilities.Available,
					},
				},
				monitor: &sacloud.MonitorLocalRouterValue{
					Time:               monitorTime,
					ReceiveBytesPerSec: 10,
					SendBytesPerSec:    20,
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "local-router",
					}),
				},
				{
					desc: c.LocalRouterInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "local-router",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.ReceiveBytesPerSec,
					metric: createGaugeWithTimestamp(10*8, map[string]string{
						"id":   "101",
						"name": "local-router",
					}, monitorTime),
				},
				{
					desc: c.SendBytesPerSec,
					metric: createGaugeWithTimestamp(20*8, map[string]string{
						"id":   "101",
						"name": "local-router",
					}, monitorTime),
				},
			},
		},
	}

	for _, tc := range cases {
		initLoggerAndErrors()
		c.logger = testLogger
		c.errors = testErrors
		c.client = tc.in

		collected, err := collectMetrics(c, "local_router")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
