package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/sakuracloud_exporter/iaas"
	"github.com/stretchr/testify/require"
)

type dummyCouponClient struct {
	coupons []*sacloud.Coupon
	err     error
}

func (d *dummyCouponClient) Find(ctx context.Context) ([]*sacloud.Coupon, error) {
	return d.coupons, d.err
}

func TestCouponCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewCouponCollector(context.Background(), testLogger, testErrors, &dummyCouponClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Discount,
		c.RemainingDays,
		c.ExpDate,
		c.Usable,
	}))
}

func TestCouponCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewCouponCollector(context.Background(), testLogger, testErrors, nil)
	untilAt := time.Now().Add(time.Hour * 24 * 3)

	cases := []struct {
		name           string
		in             iaas.CouponClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyCouponClient{
				err: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't get coupon" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyCouponClient{},
			wantMetrics: nil,
		},
		{
			name: "a coupon",
			in: &dummyCouponClient{
				coupons: []*sacloud.Coupon{
					{
						ID:         101,
						MemberID:   "memberID",
						ContractID: 201,
						Discount:   1000,
						AppliedAt:  time.Now().Add(time.Hour * -24 * 3),
						UntilAt:    untilAt,
					},
				},
			},
			wantMetrics: []*collectedMetric{
				{
					// Discount
					desc: c.Discount,
					metric: createGaugeMetric(1000, map[string]string{
						"id":          "101",
						"contract_id": "201",
						"member_id":   "memberID",
					}),
				},
				{
					// RemainingDays
					desc: c.RemainingDays,
					metric: createGaugeMetric(2, map[string]string{
						"id":          "101",
						"contract_id": "201",
						"member_id":   "memberID",
					}),
				},
				{
					// ExpirationDate
					desc: c.ExpDate,
					metric: createGaugeMetric(float64(untilAt.Unix()*1000), map[string]string{
						"id":          "101",
						"contract_id": "201",
						"member_id":   "memberID",
					}),
				},
				{
					// Usable
					desc: c.Usable,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"contract_id": "201",
						"member_id":   "memberID",
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

		collected, err := collectMetrics(c, "coupon")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
