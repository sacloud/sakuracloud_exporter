package iaas

import (
	"testing"
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/stretchr/testify/require"
)

func TestMonitor_monitorCPUTimeValue(t *testing.T) {

	cases := []struct {
		name   string
		in     []*sacloud.MonitorCPUTimeValue
		expect *sacloud.MonitorCPUTimeValue
	}{
		{
			name:   "input is nil",
			in:     nil,
			expect: nil,
		},
		{
			name: "input has only 1 value",
			in: []*sacloud.MonitorCPUTimeValue{
				{
					Time:    time.Now(),
					CPUTime: 1.0,
				},
			},
			expect: nil,
		},
		{
			name: "second last value is used: with 2 values",
			in: []*sacloud.MonitorCPUTimeValue{
				{
					Time:    time.Unix(1, 0),
					CPUTime: 1.0,
				},
				{
					Time:    time.Unix(2, 0),
					CPUTime: 2.0,
				},
			},
			expect: &sacloud.MonitorCPUTimeValue{
				Time:    time.Unix(1, 0),
				CPUTime: 1.0,
			},
		},
		{
			name: "second last value is used: with over 2 values",
			in: []*sacloud.MonitorCPUTimeValue{
				{
					Time:    time.Unix(0, 0),
					CPUTime: 0.0,
				},
				{
					Time:    time.Unix(1, 0),
					CPUTime: 1.0,
				},
				{
					Time:    time.Unix(2, 0),
					CPUTime: 2.0,
				},
			},
			expect: &sacloud.MonitorCPUTimeValue{
				Time:    time.Unix(1, 0),
				CPUTime: 1.0,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := monitorCPUTimeValue(tc.in)
			require.Equal(t, tc.expect, actual, tc.name)
		})
	}

}
