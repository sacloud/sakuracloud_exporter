// Copyright 2019-2023 The sakuracloud_exporter Authors
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
	"github.com/sacloud/iaas-api-go/helper/query"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/newsfeed"
	"github.com/sacloud/sakuracloud_exporter/platform"
	"github.com/stretchr/testify/require"
)

type dummyNFSClient struct {
	find           []*platform.NFS
	findErr        error
	monitorFree    *iaas.MonitorFreeDiskSizeValue
	monitorFreeErr error
	monitorNIC     *iaas.MonitorInterfaceValue
	monitorNICErr  error
	maintenance    *newsfeed.FeedItem
	maintenanceErr error
}

func (d *dummyNFSClient) Find(ctx context.Context) ([]*platform.NFS, error) {
	return d.find, d.findErr
}
func (d *dummyNFSClient) MonitorFreeDiskSize(ctx context.Context, zone string, id types.ID, end time.Time) (*iaas.MonitorFreeDiskSizeValue, error) {
	return d.monitorFree, d.monitorFreeErr
}
func (d *dummyNFSClient) MonitorNIC(ctx context.Context, zone string, id types.ID, end time.Time) (*iaas.MonitorInterfaceValue, error) {
	return d.monitorNIC, d.monitorNICErr
}
func (d *dummyNFSClient) MaintenanceInfo(infoURL string) (*newsfeed.FeedItem, error) {
	return d.maintenance, d.maintenanceErr
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
		c.MaintenanceScheduled,
		c.MaintenanceInfo,
		c.MaintenanceStartTime,
		c.MaintenanceEndTime,
	}))
}

func TestNFSCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewNFSCollector(context.Background(), testLogger, testErrors, nil)
	monitorTime := time.Unix(1, 0)

	cases := []struct {
		name           string
		in             platform.NFSClient
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
				find: []*platform.NFS{
					{
						ZoneName: "is1a",
						NFS: &iaas.NFS{
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
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(0, map[string]string{
						"id":   "101",
						"name": "nfs",
						"zone": "is1a",
					}),
				},
			},
		},
		{
			name: "a nfs with activity monitor",
			in: &dummyNFSClient{
				find: []*platform.NFS{
					{
						ZoneName: "is1a",
						NFS: &iaas.NFS{
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
				monitorFree: &iaas.MonitorFreeDiskSizeValue{
					Time:         monitorTime,
					FreeDiskSize: 100,
				},
				monitorNIC: &iaas.MonitorInterfaceValue{
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
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(0, map[string]string{
						"id":   "101",
						"name": "nfs",
						"zone": "is1a",
					}),
				},
			},
		},
		{
			name: "activity monitor APIs return error",
			in: &dummyNFSClient{
				find: []*platform.NFS{
					{
						ZoneName: "is1a",
						NFS: &iaas.NFS{
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
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(0, map[string]string{
						"id":   "101",
						"name": "nfs",
						"zone": "is1a",
					}),
				},
			},
			wantLogs: []string{
				`level=warn msg="can't get disk's free size: NFSID=101" err=dummy1`,
				`level=warn msg="can't get nfs's NIC metrics: NFSID=101" err=dummy2`,
			},
			wantErrCounter: 2,
		},
		{
			name: "a nfs with maintenance info",
			in: &dummyNFSClient{
				find: []*platform.NFS{
					{
						ZoneName: "is1a",
						NFS: &iaas.NFS{
							ID:                  101,
							Name:                "nfs",
							Tags:                types.Tags{"tag1", "tag2"},
							Description:         "desc",
							InstanceHostName:    "sacXXX",
							InstanceStatus:      types.ServerInstanceStatuses.Up,
							Availability:        types.Availabilities.Available,
							InstanceHostInfoURL: "http://example.com/maintenance-info-dummy-url",
							IPAddresses:         []string{"192.168.0.11"},
							DefaultRoute:        "192.168.0.1",
							NetworkMaskLen:      24,
							SwitchID:            201,
							SwitchName:          "switch",
						},
						Plan: &query.NFSPlanInfo{
							NFSPlanID:  1001,
							Size:       types.NFSHDDSizes.Size100GB,
							DiskPlanID: types.NFSPlans.HDD,
						},
						PlanName: "HDD 100GB",
					},
				},
				maintenance: &newsfeed.FeedItem{
					StrDate:       "947430000", // 2000-01-10
					Description:   "desc",
					StrEventStart: "946652400", // 2000-01-01
					StrEventEnd:   "949244400", // 2000-01-31
					Title:         "dummy-title",
					URL:           "http://example.com/maintenance",
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
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "nfs",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "nfs",
						"zone":        "is1a",
						"info_url":    "http://example.com/maintenance",
						"info_title":  "dummy-title",
						"description": "desc",
						"start_date":  "946652400",
						"end_date":    "949244400",
					}),
				},
				{
					desc: c.MaintenanceStartTime,
					metric: createGaugeMetric(946652400, map[string]string{
						"id":   "101",
						"name": "nfs",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceEndTime,
					metric: createGaugeMetric(949244400, map[string]string{
						"id":   "101",
						"name": "nfs",
						"zone": "is1a",
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

		collected, err := collectMetrics(c, "nfs")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
