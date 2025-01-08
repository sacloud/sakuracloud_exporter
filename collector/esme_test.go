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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/sakuracloud_exporter/platform"
	"github.com/stretchr/testify/require"
)

type dummyESMEClient struct {
	esme    []*iaas.ESME
	findErr error
	logs    []*iaas.ESMELogs
	logsErr error
}

func (d *dummyESMEClient) Find(ctx context.Context) ([]*iaas.ESME, error) {
	return d.esme, d.findErr
}

func (d *dummyESMEClient) Logs(ctx context.Context, esmeID types.ID) ([]*iaas.ESMELogs, error) {
	return d.logs, d.logsErr
}

func TestESMECollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewESMECollector(context.Background(), testLogger, testErrors, &dummyESMEClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.ESMEInfo,
		c.MessageCount,
	}))
}

func TestESMECollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewESMECollector(context.Background(), testLogger, testErrors, nil)

	cases := []struct {
		name           string
		in             platform.ESMEClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyESMEClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=WARN msg="can't list ESME" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyESMEClient{},
			wantMetrics: nil,
		},
		{
			name: "esme: collecting ESME logs is failed ",
			in: &dummyESMEClient{
				esme: []*iaas.ESME{
					{
						ID:          101,
						Name:        "ESME",
						Tags:        types.Tags{"tag1", "tag2"},
						Description: "desc",
					},
				},
				logsErr: errors.New("dummy"),
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.ESMEInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "ESME",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
			},
			wantLogs:       []string{`level=WARN msg="can't collect logs of the esme[101]" err=dummy`},
			wantErrCounter: 1,
		},
	}

	for _, tc := range cases {
		initLoggerAndErrors()
		c.logger = testLogger
		c.errors = testErrors
		c.client = tc.in

		collected, err := collectMetrics(c, "esme")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
