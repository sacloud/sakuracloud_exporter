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

type dummySIMClient struct {
	find         []*sacloud.SIM
	findErr      error
	nopConfig    []*sacloud.SIMNetworkOperatorConfig
	nopConfigErr error
	monitor      *sacloud.MonitorLinkValue
	monitorErr   error
}

func (d *dummySIMClient) Find(ctx context.Context) ([]*sacloud.SIM, error) {
	return d.find, d.findErr
}
func (d *dummySIMClient) GetNetworkOperatorConfig(ctx context.Context, id types.ID) ([]*sacloud.SIMNetworkOperatorConfig, error) {
	return d.nopConfig, d.nopConfigErr
}
func (d *dummySIMClient) MonitorTraffic(ctx context.Context, id types.ID, end time.Time) (*sacloud.MonitorLinkValue, error) {
	return d.monitor, d.monitorErr
}

func TestSIMCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewSIMCollector(context.Background(), testLogger, testErrors, &dummySIMClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Up,
		c.SIMInfo,
		c.Uplink,
		c.Downlink,
	}))
}

func TestSIMCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewSIMCollector(context.Background(), testLogger, testErrors, nil)
	monitorTime := time.Unix(1, 0)

	cases := []struct {
		name           string
		in             iaas.SIMClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummySIMClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't list sims" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummySIMClient{},
			wantMetrics: nil,
		},
		{
			name: "a SIM with activity monitor",
			in: &dummySIMClient{
				find: []*sacloud.SIM{
					{
						ID:   101,
						Name: "sim",
						Info: &sacloud.SIMInfo{
							IMEILock:       true,
							RegisteredDate: time.Unix(1, 0),
							ActivatedDate:  time.Unix(2, 0),
							//DeactivatedDate: time.Unix(3, 0),
							IP:            "192.0.2.1",
							SIMGroupID:    "201",
							SessionStatus: "UP",
							TrafficBytesOfCurrentMonth: &sacloud.SIMTrafficBytes{
								UplinkBytes:   100 * 1000,
								DownlinkBytes: 200 * 1000,
							},
						},
						Tags:        types.Tags{"tag1", "tag2"},
						Description: "desc",
					},
				},
				nopConfig: []*sacloud.SIMNetworkOperatorConfig{
					{Allow: true, Name: "docomo"},
					{Allow: false, Name: "softbank"},
					{Allow: true, Name: "kddi"},
				},
				monitor: &sacloud.MonitorLinkValue{
					Time:        monitorTime,
					UplinkBPS:   10 * 1000,
					DownlinkBPS: 20 * 1000,
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "sim",
					}),
				},
				{
					desc: c.SIMInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":               "101",
						"name":             "sim",
						"imei_lock":        "1",
						"registerd_date":   "1000",
						"activated_date":   "2000",
						"deactivated_date": "0",
						"ipaddress":        "192.0.2.1",
						"simgroup_id":      "201",
						"carriers":         ",docomo,kddi,",
						"tags":             ",tag1,tag2,",
						"description":      "desc",
					}),
				},
				{
					desc: c.Uplink,
					metric: createGaugeWithTimestamp(10, map[string]string{
						"id":   "101",
						"name": "sim",
					}, monitorTime),
				},
				{
					desc: c.Downlink,
					metric: createGaugeWithTimestamp(20, map[string]string{
						"id":   "101",
						"name": "sim",
					}, monitorTime),
				},
			},
		},
		{
			name: "APIs return error",
			in: &dummySIMClient{
				find: []*sacloud.SIM{
					{
						ID:   101,
						Name: "sim",
						Info: &sacloud.SIMInfo{
							IMEILock:       true,
							RegisteredDate: time.Unix(1, 0),
							ActivatedDate:  time.Unix(2, 0),
							//DeactivatedDate: time.Unix(3, 0),
							IP:            "192.0.2.1",
							SIMGroupID:    "201",
							SessionStatus: "UP",
						},
						Tags:        types.Tags{"tag1", "tag2"},
						Description: "desc",
					},
				},
				nopConfigErr: errors.New("dummy1"),
				monitorErr:   errors.New("dummy2"),
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "sim",
					}),
				},
			},
			wantErrCounter: 2,
			wantLogs: []string{
				`level=warn msg="can't get sim's metrics: SIMID=101" err=dummy2`,
				`level=warn msg="can't get sim's network operator config: SIMID=101" err=dummy1`,
			},
		},
	}

	for _, tc := range cases {
		initLoggerAndErrors()
		c.logger = testLogger
		c.errors = testErrors
		c.client = tc.in

		collected, err := collectMetrics(c, "sim")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
