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
	"github.com/sacloud/libsacloud/v2/helper/query"
	"github.com/sacloud/sakuracloud_exporter/iaas"
	"github.com/stretchr/testify/require"
)

type dummyNFSClient struct {
	find           []*iaas.NFS
	findErr        error
	monitorFree    *sacloud.MonitorFreeDiskSizeValue
	monitorFreeErr error
	monitorNIC     *sacloud.MonitorInterfaceValue
	monitorNICErr  error
}

func (d *dummyNFSClient) Find(ctx context.Context) ([]*iaas.NFS, error) {
	return d.find, d.findErr
}
func (d *dummyNFSClient) MonitorFreeDiskSize(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorFreeDiskSizeValue, error) {
	return d.monitorFree, d.monitorFreeErr
}
func (d *dummyNFSClient) MonitorNIC(ctx context.Context, zone string, id types.ID, end time.Time) (*sacloud.MonitorInterfaceValue, error) {
	return d.monitorNIC, d.monitorNICErr
}

func TestNFSCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewNFSCollector(context.Background(), testLogger, testErrors, &dummyNFSClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Up,
		c.NFSInfo,
		c.DiskFree,
		c.NICInfo,
		c.NICReceive,
		c.NICSend,
	}))
}

func TestNFSCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewNFSCollector(context.Background(), testLogger, testErrors, nil)
	monitorTime := time.Unix(1, 0)

	cases := []struct {
		name           string
		in             iaas.NFSClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyNFSClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't list nfs" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyNFSClient{},
			wantMetrics: nil,
		},
		{
			name: "a nfs without activity monitor",
			in: &dummyNFSClient{
				find: []*iaas.NFS{
					{
						ZoneName: "is1a",
						NFS: &sacloud.NFS{
							ID:               101,
							Name:             "nfs",
							Tags:             types.Tags{"tag1", "tag2"},
							Description:      "desc",
							InstanceHostName: "sacXXX",
							InstanceStatus:   types.ServerInstanceStatuses.Up,
							IPAddresses:      []string{"192.168.0.11"},
							DefaultRoute:     "192.168.0.1",
							NetworkMaskLen:   24,
							SwitchID:         201,
							SwitchName:       "switch",
						},
						Plan: &query.NFSPlanInfo{
							NFSPlanID:  1001,
							Size:       types.NFSHDDSizes.Size100GB,
							DiskPlanID: types.NFSPlans.HDD,
						},
						PlanName: "HDD 100GB",
					},
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "nfs",
						"zone": "is1a",
					}),
				},
				{
					desc: c.NFSInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "nfs",
						"zone":        "is1a",
						"plan":        "HDD 100GB",
						"size":        "100",
						"host":        "sacXXX",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.NICInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":            "101",
						"name":          "nfs",
						"zone":          "is1a",
						"upstream_id":   "201",
						"upstream_name": "switch",
						"ipaddress":     "192.168.0.11",
						"nw_mask_len":   "24",
						"gateway":       "192.168.0.1",
					}),
				},
			},
		},
		{
			name: "a nfs with activity monitor",
			in: &dummyNFSClient{
				find: []*iaas.NFS{
					{
						ZoneName: "is1a",
						NFS: &sacloud.NFS{
							ID:               101,
							Name:             "nfs",
							Tags:             types.Tags{"tag1", "tag2"},
							Description:      "desc",
							InstanceHostName: "sacXXX",
							InstanceStatus:   types.ServerInstanceStatuses.Up,
							Availability:     types.Availabilities.Available,
							IPAddresses:      []string{"192.168.0.11"},
							DefaultRoute:     "192.168.0.1",
							NetworkMaskLen:   24,
							SwitchID:         201,
							SwitchName:       "switch",
						},
						Plan: &query.NFSPlanInfo{
							NFSPlanID:  1001,
							Size:       types.NFSHDDSizes.Size100GB,
							DiskPlanID: types.NFSPlans.HDD,
						},
						PlanName: "HDD 100GB",
					},
				},
				monitorFree: &sacloud.MonitorFreeDiskSizeValue{
					Time:         monitorTime,
					FreeDiskSize: 100,
				},
				monitorNIC: &sacloud.MonitorInterfaceValue{
					Time:    monitorTime,
					Receive: 200,
					Send:    300,
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "nfs",
						"zone": "is1a",
					}),
				},
				{
					desc: c.NFSInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "nfs",
						"zone":        "is1a",
						"plan":        "HDD 100GB",
						"size":        "100",
						"host":        "sacXXX",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.NICInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":            "101",
						"name":          "nfs",
						"zone":          "is1a",
						"upstream_id":   "201",
						"upstream_name": "switch",
						"ipaddress":     "192.168.0.11",
						"nw_mask_len":   "24",
						"gateway":       "192.168.0.1",
					}),
				},
				{
					desc: c.DiskFree,
					metric: createGaugeWithTimestamp(float64(100)/1024/1024, map[string]string{
						"id":   "101",
						"name": "nfs",
						"zone": "is1a",
					}, monitorTime),
				},
				{
					desc: c.NICReceive,
					metric: createGaugeWithTimestamp(float64(200)*8/1000, map[string]string{
						"id":   "101",
						"name": "nfs",
						"zone": "is1a",
					}, monitorTime),
				},
				{
					desc: c.NICSend,
					metric: createGaugeWithTimestamp(float64(300)*8/1000, map[string]string{
						"id":   "101",
						"name": "nfs",
						"zone": "is1a",
					}, monitorTime),
				},
			},
		},
		{
			name: "activity monitor APIs return error",
			in: &dummyNFSClient{
				find: []*iaas.NFS{
					{
						ZoneName: "is1a",
						NFS: &sacloud.NFS{
							ID:               101,
							Name:             "nfs",
							Tags:             types.Tags{"tag1", "tag2"},
							Description:      "desc",
							InstanceHostName: "sacXXX",
							InstanceStatus:   types.ServerInstanceStatuses.Up,
							Availability:     types.Availabilities.Available,
							IPAddresses:      []string{"192.168.0.11"},
							DefaultRoute:     "192.168.0.1",
							NetworkMaskLen:   24,
							SwitchID:         201,
							SwitchName:       "switch",
						},
						Plan: &query.NFSPlanInfo{
							NFSPlanID:  1001,
							Size:       types.NFSHDDSizes.Size100GB,
							DiskPlanID: types.NFSPlans.HDD,
						},
						PlanName: "HDD 100GB",
					},
				},
				monitorFreeErr: errors.New("dummy1"),
				monitorNICErr:  errors.New("dummy2"),
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "nfs",
						"zone": "is1a",
					}),
				},
				{
					desc: c.NFSInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "nfs",
						"zone":        "is1a",
						"plan":        "HDD 100GB",
						"size":        "100",
						"host":        "sacXXX",
						"tags":        ",tag1,tag2,",
						"description": "desc",
					}),
				},
				{
					desc: c.NICInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":            "101",
						"name":          "nfs",
						"zone":          "is1a",
						"upstream_id":   "201",
						"upstream_name": "switch",
						"ipaddress":     "192.168.0.11",
						"nw_mask_len":   "24",
						"gateway":       "192.168.0.1",
					}),
				},
			},
			wantLogs: []string{
				`level=warn msg="can't get disk's free size: NFSID=101" err=dummy1`,
				`level=warn msg="can't get nfs's NIC metrics: NFSID=101" err=dummy2`,
			},
			wantErrCounter: 2,
		},
	}

	for _, tc := range cases {
		initLoggerAndErrors()
		c.logger = testLogger
		c.errors = testErrors
		c.client = tc.in

		collected, err := collectMetrics(c, "nfs")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
