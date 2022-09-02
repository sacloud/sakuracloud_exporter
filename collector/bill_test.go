// Copyright 2019-2022 The sakuracloud_exporter Authors
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
	"github.com/sacloud/sakuracloud_exporter/platform"
	"github.com/stretchr/testify/require"
)

type dummyBillClient struct {
	bill *iaas.Bill
	err  error
}

func (d *dummyBillClient) Read(ctx context.Context) (*iaas.Bill, error) {
	return d.bill, d.err
}

func TestBillCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewBillCollector(context.Background(), testLogger, testErrors, &dummyBillClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Amount,
	}))
}

func TestBillCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewBillCollector(context.Background(), testLogger, testErrors, nil)

	cases := []struct {
		name           string
		in             platform.BillClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyBillClient{
				err: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't get bill" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyBillClient{},
			wantMetrics: nil,
		},
		{
			name: "a bill",
			in: &dummyBillClient{
				bill: &iaas.Bill{
					ID:             101,
					Amount:         1234,
					Date:           time.Now(),
					MemberID:       "memberID",
					Paid:           false,
					PayLimit:       time.Now(),
					PaymentClassID: 0,
				},
			},
			wantMetrics: []*collectedMetric{
				{
					// Discount
					desc: c.Amount,
					metric: createGaugeMetric(1234, map[string]string{
						"member_id": "memberID",
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

		collected, err := collectMetrics(c, "bill")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
