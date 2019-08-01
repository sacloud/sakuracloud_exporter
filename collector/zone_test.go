package collector

import (
	"context"
	"errors"
	"testing"

	dto "github.com/prometheus/client_model/go"

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
	zoneCollector := NewZoneCollector(context.Background(), testLogger, testErrors, &dummyZoneClient{})

	descs := collectDescs(zoneCollector)
	require.Len(t, descs, 1)
}

func TestZoneCollector_Collect(t *testing.T) {
	cases := []struct {
		name           string
		in             iaas.ZoneClient
		wantLog        string
		wantErrCounter float64
		wantMetrics    []*dto.Metric
	}{
		{
			name: "collector returns error",
			in: &dummyZoneClient{
				err: errors.New("dummy"),
			},
			wantLog:        `level=warn msg="can't get zone info" err=dummy`,
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:           "empty result",
			in:             &dummyZoneClient{},
			wantLog:        "",
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
			wantLog:        "",
			wantErrCounter: 0,
			wantMetrics: []*dto.Metric{
				createGaugeMetric(1, map[string]string{
					"id":          "1",
					"name":        "zone",
					"description": "desc",
					"region_id":   "2",
					"region_name": "region",
				}),
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
			wantLog:        "",
			wantErrCounter: 0,
			wantMetrics: []*dto.Metric{
				createGaugeMetric(1, map[string]string{
					"id":          "1",
					"name":        "zone1",
					"description": "desc1",
					"region_id":   "2",
					"region_name": "region2",
				}),
				createGaugeMetric(1, map[string]string{
					"id":          "3",
					"name":        "zone3",
					"description": "desc3",
					"region_id":   "4",
					"region_name": "region4",
				}),
			},
		},
	}

	for _, tc := range cases {
		initLoggerAndErrors()
		collector := NewZoneCollector(context.Background(), testLogger, testErrors, tc.in)
		collected, err := collectMetrics(collector, "zone")
		require.NoError(t, err)
		require.Equal(t, tc.wantLog, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		require.Equal(t, tc.wantMetrics, collected.collected)
	}
}
