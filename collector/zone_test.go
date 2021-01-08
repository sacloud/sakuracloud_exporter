// Copyright 2019-2021 The sakuracloud_exporter Authors
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

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/sakuracloud_exporter/iaas"
	"github.com/stretchr/testify/require"
)

type dummyZoneClient struct {
	zones []*sacloud.Zone
	err   error
}

func (d *dummyZoneClient) Find(ctx context.Context) ([]*sacloud.Zone, error) {
	return d.zones, d.err
}

func TestZoneCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewZoneCollector(context.Background(), testLogger, testErrors, &dummyZoneClient{})

	descs := collectDescs(c)
	require.Len(t, descs, 1)
}

func TestZoneCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewZoneCollector(context.Background(), testLogger, testErrors, nil)

	cases := []struct {
		name           string
		in             iaas.ZoneClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyZoneClient{
				err: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't get zone info" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:           "empty result",
			in:             &dummyZoneClient{},
			wantLogs:       nil,
			wantErrCounter: 0,
			wantMetrics:    nil,
		},
		{
			name: "with single zone info",
			in: &dummyZoneClient{
				zones: []*sacloud.Zone{
					{
						ID:          1,
						Name:        "zone",
						Description: "desc",
						Region: &sacloud.Region{
							ID:   2,
							Name: "region",
						},
					},
				},
			},
			wantLogs:       nil,
			wantErrCounter: 0,
			wantMetrics: []*collectedMetric{
				{
					desc: c.ZoneInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "1",
						"name":        "zone",
						"description": "desc",
						"region_id":   "2",
						"region_name": "region",
					}),
				},
			},
		},
		{
			name: "with multiple zone info",
			in: &dummyZoneClient{
				zones: []*sacloud.Zone{
					{
						ID:          1,
						Name:        "zone1",
						Description: "desc1",
						Region: &sacloud.Region{
							ID:   2,
							Name: "region2",
						},
					},
					{
						ID:          3,
						Name:        "zone3",
						Description: "desc3",
						Region: &sacloud.Region{
							ID:   4,
							Name: "region4",
						},
					},
				},
			},
			wantLogs:       nil,
			wantErrCounter: 0,
			wantMetrics: []*collectedMetric{
				{
					desc: c.ZoneInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "1",
						"name":        "zone1",
						"description": "desc1",
						"region_id":   "2",
						"region_name": "region2",
					}),
				},
				{
					desc: c.ZoneInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "3",
						"name":        "zone3",
						"description": "desc3",
						"region_id":   "4",
						"region_name": "region4",
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

		collected, err := collectMetrics(c, "zone")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		require.Equal(t, tc.wantMetrics, collected.collected)
	}
}
