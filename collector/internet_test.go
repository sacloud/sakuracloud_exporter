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
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/sakuracloud_exporter/platform"
	"github.com/stretchr/testify/require"
)

type dummyInternetClient struct {
	find       []*platform.Internet
	findErr    error
	monitor    *iaas.MonitorRouterValue
	monitorErr error
}

func (d *dummyInternetClient) Find(ctx context.Context) ([]*platform.Internet, error) {
	return d.find, d.findErr
}

func (d *dummyInternetClient) MonitorTraffic(ctx context.Context, zone string, internetID types.ID, end time.Time) (*iaas.MonitorRouterValue, error) {
	return d.monitor, d.monitorErr
}

func TestInternetCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewInternetCollector(context.Background(), testLogger, testErrors, &dummyInternetClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Info,
		c.In,
		c.Out,
	}))
}

func TestInternetCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewInternetCollector(context.Background(), testLogger, testErrors, nil)
	monitorTime := time.Unix(1, 0)

	cases := []struct {
		name           string
		in             platform.InternetClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyInternetClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=WARN msg="can't list internets" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyInternetClient{},
			wantMetrics: nil,
		},
		{
			name: "a internet router",
			in: &dummyInternetClient{
				find: []*platform.Internet{
					{
						ZoneName: "is1a",
						Internet: &iaas.Internet{
							ID:          101,
							Name:        "internet",
							Description: "desc",
							Tags:        types.Tags{"tag1", "tag2"},
							Switch: &iaas.SwitchInfo{
								ID:   201,
								Name: "switch",
							},
							BandWidthMbps: 100,
						},
					},
				},
				monitor: &iaas.MonitorRouterValue{
					Time: monitorTime,
					In:   1000,
					Out:  2000,
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Info,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "internet",
						"zone":        "is1a",
						"switch_id":   "201",
						"bandwidth":   "100",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.In,
					metric: createGaugeWithTimestamp(1, map[string]string{
						"id":        "101",
						"name":      "internet",
						"zone":      "is1a",
						"switch_id": "201",
					}, monitorTime),
				},
				{
					desc: c.Out,
					metric: createGaugeWithTimestamp(2, map[string]string{
						"id":        "101",
						"name":      "internet",
						"zone":      "is1a",
						"switch_id": "201",
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

		collected, err := collectMetrics(c, "internet")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
