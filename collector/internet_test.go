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

type dummyInternetClient struct {
	find       []*iaas.Internet
	findErr    error
	monitor    *sacloud.MonitorRouterValue
	monitorErr error
}

func (d *dummyInternetClient) Find(ctx context.Context) ([]*iaas.Internet, error) {
	return d.find, d.findErr
}

func (d *dummyInternetClient) MonitorTraffic(ctx context.Context, zone string, internetID types.ID, end time.Time) (*sacloud.MonitorRouterValue, error) {
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
		in             iaas.InternetClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyInternetClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't list internets" err=dummy`},
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
				find: []*iaas.Internet{
					{
						ZoneName: "is1a",
						Internet: &sacloud.Internet{
							ID:          101,
							Name:        "internet",
							Description: "desc",
							Tags:        types.Tags{"tag1", "tag2"},
							Switch: &sacloud.SwitchInfo{
								ID:   201,
								Name: "switch",
							},
							BandWidthMbps: 100,
						},
					},
				},
				monitor: &sacloud.MonitorRouterValue{
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
