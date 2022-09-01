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
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/newsfeed"
	"github.com/sacloud/sakuracloud_exporter/platform"
	"github.com/stretchr/testify/require"
)

type dummyServerClient struct {
	find           []*platform.Server
	findErr        error
	readDisk       *iaas.Disk
	readDiskErr    error
	monitorCPU     *iaas.MonitorCPUTimeValue
	monitorCPUErr  error
	monitorDisk    *iaas.MonitorDiskValue
	monitorDiskErr error
	monitorNIC     *iaas.MonitorInterfaceValue
	monitorNICErr  error
	maintenance    *newsfeed.FeedItem
	maintenanceErr error
}

func (d *dummyServerClient) Find(ctx context.Context) ([]*platform.Server, error) {
	return d.find, d.findErr
}

func (d *dummyServerClient) ReadDisk(ctx context.Context, zone string, diskID types.ID) (*iaas.Disk, error) {
	return d.readDisk, d.readDiskErr
}

func (d *dummyServerClient) MonitorCPU(ctx context.Context, zone string, id types.ID, end time.Time) (*iaas.MonitorCPUTimeValue, error) {
	return d.monitorCPU, d.monitorCPUErr
}
func (d *dummyServerClient) MonitorDisk(ctx context.Context, zone string, diskID types.ID, end time.Time) (*iaas.MonitorDiskValue, error) {
	return d.monitorDisk, d.monitorDiskErr
}
func (d *dummyServerClient) MonitorNIC(ctx context.Context, zone string, nicID types.ID, end time.Time) (*iaas.MonitorInterfaceValue, error) {
	return d.monitorNIC, d.monitorNICErr
}
func (d *dummyServerClient) MaintenanceInfo(infoURL string) (*newsfeed.FeedItem, error) {
	return d.maintenance, d.maintenanceErr
}

func TestServerCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewServerCollector(context.Background(), testLogger, testErrors, &dummyServerClient{}, false)

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Up,
		c.ServerInfo,
		c.CPUs,
		c.CPUTime,
		c.Memories,
		c.DiskInfo,
		c.DiskRead,
		c.DiskWrite,
		c.NICInfo,
		c.NICBandwidth,
		c.NICReceive,
		c.NICSend,
	}))
}

func TestServerCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewServerCollector(context.Background(), testLogger, testErrors, nil, false)
	monitorTime := time.Unix(1, 0)

	server := &platform.Server{
		ZoneName: "is1a",
		Server: &iaas.Server{
			ID:               101,
			Name:             "server",
			Description:      "desc",
			Tags:             types.Tags{"tag1", "tag2"},
			CPU:              2,
			MemoryMB:         4 * 1024,
			InstanceStatus:   types.ServerInstanceStatuses.Up,
			Availability:     types.Availabilities.Available,
			InstanceHostName: "sacXXX",
			Disks: []*iaas.ServerConnectedDisk{
				{
					ID:         201,
					Name:       "disk",
					DiskPlanID: types.DiskPlans.SSD,
					Connection: types.DiskConnections.VirtIO,
					SizeMB:     20 * 1024,
					Storage: &iaas.Storage{
						ID:         1001,
						Class:      "iscsi1204",
						Generation: 100,
					},
				},
			},
			PrivateHostID:   3001,
			PrivateHostName: "private-host-name",
			Interfaces: []*iaas.InterfaceView{
				{
					ID:           301,
					SwitchID:     401,
					SwitchName:   "switch",
					UpstreamType: types.UpstreamNetworkTypes.Switch,
				},
			},
		},
	}

	cases := []struct {
		name           string
		in             platform.ServerClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyServerClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't list servers" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyServerClient{},
			wantMetrics: nil,
		},
		{
			name: "a server with activity monitors",
			in: &dummyServerClient{
				find: []*platform.Server{server},
				monitorCPU: &iaas.MonitorCPUTimeValue{
					Time:    monitorTime,
					CPUTime: 100,
				},
				monitorDisk: &iaas.MonitorDiskValue{
					Time:  monitorTime,
					Read:  201,
					Write: 202,
				},
				monitorNIC: &iaas.MonitorInterfaceValue{
					Time:    monitorTime,
					Receive: 301,
					Send:    302,
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.ServerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":              "101",
						"name":            "server",
						"zone":            "is1a",
						"cpus":            "2",
						"disks":           "1",
						"nics":            "1",
						"memories":        "4",
						"host":            "sacXXX",
						"tags":            ",tag1,tag2,",
						"description":     "desc",
						"private_host_id": "3001",
					}),
				},
				{
					desc: c.CPUs,
					metric: createGaugeMetric(2, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.Memories,
					metric: createGaugeMetric(4, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(0, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				//{
				//	desc: c.DiskInfo,
				//	metric: createGaugeMetric(1, map[string]string{
				//		"id":                 "101",
				//		"name":               "server",
				//		"zone":               "is1a",
				//		"disk_id":            "201",
				//		"disk_name":          "disk",
				//		"index":              "0",
				//		"plan":               "ssd",
				//		"interface":          "virtio",
				//		"size":               "20",
				//		"tags":               ",disk1,disk2,",
				//		"description":        "disk-desc",
				//		"storage_id":         "1001",
				//		"storage_class":      "iscsi1204",
				//		"storage_generation": "100",
				//	}),
				// },
				{
					desc: c.NICBandwidth,
					metric: createGaugeMetric(0, map[string]string{ // 専有ホストの場合は0
						"id":           "101",
						"name":         "server",
						"zone":         "is1a",
						"index":        "0",
						"interface_id": "301",
					}),
				},
				{
					desc: c.NICInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":            "101",
						"name":          "server",
						"zone":          "is1a",
						"index":         "0",
						"interface_id":  "301",
						"upstream_id":   "401",
						"upstream_name": "switch",
						"upstream_type": "switch",
					}),
				},
				{
					desc: c.CPUTime,
					metric: createGaugeWithTimestamp(float64(100)*1000, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}, monitorTime),
				},
				{
					desc: c.DiskRead,
					metric: createGaugeWithTimestamp(float64(201)/1024, map[string]string{
						"id":        "101",
						"name":      "server",
						"zone":      "is1a",
						"disk_id":   "201",
						"disk_name": "disk",
						"index":     "0",
					}, monitorTime),
				},
				{
					desc: c.DiskWrite,
					metric: createGaugeWithTimestamp(float64(202)/1024, map[string]string{
						"id":        "101",
						"name":      "server",
						"zone":      "is1a",
						"disk_id":   "201",
						"disk_name": "disk",
						"index":     "0",
					}, monitorTime),
				},
				{
					desc: c.NICReceive,
					metric: createGaugeWithTimestamp(float64(301)*8/1000, map[string]string{
						"id":           "101",
						"name":         "server",
						"zone":         "is1a",
						"index":        "0",
						"interface_id": "301",
					}, monitorTime),
				},
				{
					desc: c.NICSend,
					metric: createGaugeWithTimestamp(float64(302)*8/1000, map[string]string{
						"id":           "101",
						"name":         "server",
						"zone":         "is1a",
						"index":        "0",
						"interface_id": "301",
					}, monitorTime),
				},
			},
		},
		{
			name: "activity monitor APIs return error",
			in: &dummyServerClient{
				find:           []*platform.Server{server},
				monitorCPUErr:  errors.New("dummy1"),
				monitorDiskErr: errors.New("dummy2"),
				monitorNICErr:  errors.New("dummy3"),
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.ServerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":              "101",
						"name":            "server",
						"zone":            "is1a",
						"cpus":            "2",
						"disks":           "1",
						"nics":            "1",
						"memories":        "4",
						"host":            "sacXXX",
						"tags":            ",tag1,tag2,",
						"description":     "desc",
						"private_host_id": "3001",
					}),
				},
				{
					desc: c.CPUs,
					metric: createGaugeMetric(2, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.Memories,
					metric: createGaugeMetric(4, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(0, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				//{
				//	desc: c.DiskInfo,
				//	metric: createGaugeMetric(1, map[string]string{
				//		"id":                 "101",
				//		"name":               "server",
				//		"zone":               "is1a",
				//		"disk_id":            "201",
				//		"disk_name":          "disk",
				//		"index":              "0",
				//		"plan":               "ssd",
				//		"interface":          "virtio",
				//		"size":               "20",
				//		"tags":               ",disk1,disk2,",
				//		"description":        "disk-desc",
				//		"storage_id":         "1001",
				//		"storage_class":      "iscsi1204",
				//		"storage_generation": "100",
				//	}),
				// },
				{
					desc: c.NICBandwidth,
					metric: createGaugeMetric(0, map[string]string{
						"id":           "101",
						"name":         "server",
						"zone":         "is1a",
						"index":        "0",
						"interface_id": "301",
					}),
				},
				{
					desc: c.NICInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":            "101",
						"name":          "server",
						"zone":          "is1a",
						"index":         "0",
						"interface_id":  "301",
						"upstream_id":   "401",
						"upstream_name": "switch",
						"upstream_type": "switch",
					}),
				},
			},
			wantErrCounter: 3,
			wantLogs: []string{
				`level=warn msg="can't get disk's metrics: ServerID=101, DiskID=201" err=dummy2`,
				`level=warn msg="can't get nic's metrics: ServerID=101,NICID=301" err=dummy3`,
				`level=warn msg="can't get server's CPU-TIME: ID=101" err=dummy1`,
			},
		},
		{
			name: "maintenance info",
			in: &dummyServerClient{
				find: []*platform.Server{
					{
						ZoneName: "is1a",
						Server: &iaas.Server{
							ID:                  101,
							Name:                "server",
							CPU:                 2,
							MemoryMB:            4 * 1024,
							InstanceStatus:      types.ServerInstanceStatuses.Up,
							Availability:        types.Availabilities.Available,
							InstanceHostName:    "sacXXX",
							InstanceHostInfoURL: "https://maintenance.example.com",
						},
					},
				},
				maintenance: &newsfeed.FeedItem{
					StrDate:       fmt.Sprintf("%d", time.Unix(1, 0).Unix()),
					Description:   "maintenance-desc",
					StrEventStart: fmt.Sprintf("%d", time.Unix(2, 0).Unix()),
					StrEventEnd:   fmt.Sprintf("%d", time.Unix(3, 0).Unix()),
					Title:         "maintenance-title",
					URL:           "https://maintenance.example.com/?entry=1",
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Up,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.ServerInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":              "101",
						"name":            "server",
						"zone":            "is1a",
						"cpus":            "2",
						"disks":           "0",
						"nics":            "0",
						"memories":        "4",
						"host":            "sacXXX",
						"tags":            "",
						"description":     "",
						"private_host_id": "",
					}),
				},
				{
					desc: c.CPUs,
					metric: createGaugeMetric(2, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.Memories,
					metric: createGaugeMetric(4, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "server",
						"zone":        "is1a",
						"info_url":    "https://maintenance.example.com/?entry=1",
						"info_title":  "maintenance-title",
						"description": "maintenance-desc",
						"start_date":  fmt.Sprintf("%d", time.Unix(2, 0).Unix()),
						"end_date":    fmt.Sprintf("%d", time.Unix(3, 0).Unix()),
					}),
				},
				{
					desc: c.MaintenanceStartTime,
					metric: createGaugeMetric(float64(time.Unix(2, 0).Unix()), map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceEndTime,
					metric: createGaugeMetric(float64(time.Unix(3, 0).Unix()), map[string]string{
						"id":   "101",
						"name": "server",
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

		collected, err := collectMetrics(c, "server")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}

func TestServerCollector_CollectMaintenanceOnly(t *testing.T) {
	initLoggerAndErrors()
	c := NewServerCollector(context.Background(), testLogger, testErrors, nil, true)
	monitorTime := time.Unix(1, 0)

	server := &platform.Server{
		ZoneName: "is1a",
		Server: &iaas.Server{
			ID:               101,
			Name:             "server",
			Description:      "desc",
			Tags:             types.Tags{"tag1", "tag2"},
			CPU:              2,
			MemoryMB:         4 * 1024,
			InstanceStatus:   types.ServerInstanceStatuses.Up,
			Availability:     types.Availabilities.Available,
			InstanceHostName: "sacXXX",
			Disks: []*iaas.ServerConnectedDisk{
				{
					ID:         201,
					Name:       "disk",
					DiskPlanID: types.DiskPlans.SSD,
					Connection: types.DiskConnections.VirtIO,
					SizeMB:     20 * 1024,
					Storage: &iaas.Storage{
						ID:         1001,
						Class:      "iscsi1204",
						Generation: 100,
					},
				},
			},
			Interfaces: []*iaas.InterfaceView{
				{
					ID:           301,
					SwitchID:     401,
					SwitchName:   "switch",
					UpstreamType: types.UpstreamNetworkTypes.Switch,
				},
			},
		},
	}

	cases := []struct {
		name           string
		in             platform.ServerClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyServerClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't list servers" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyServerClient{},
			wantMetrics: nil,
		},
		{
			name: "a server maintenance scheduled",
			in: &dummyServerClient{
				find: []*platform.Server{server},
				monitorCPU: &iaas.MonitorCPUTimeValue{
					Time:    monitorTime,
					CPUTime: 100,
				},
				monitorDisk: &iaas.MonitorDiskValue{
					Time:  monitorTime,
					Read:  201,
					Write: 202,
				},
				monitorNIC: &iaas.MonitorInterfaceValue{
					Time:    monitorTime,
					Receive: 301,
					Send:    302,
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(0, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
			},
		},
		{
			name: "maintenance info",
			in: &dummyServerClient{
				find: []*platform.Server{
					{
						ZoneName: "is1a",
						Server: &iaas.Server{
							ID:                  101,
							Name:                "server",
							CPU:                 2,
							MemoryMB:            4 * 1024,
							InstanceStatus:      types.ServerInstanceStatuses.Up,
							Availability:        types.Availabilities.Available,
							InstanceHostName:    "sacXXX",
							InstanceHostInfoURL: "https://maintenance.example.com",
						},
					},
				},
				maintenance: &newsfeed.FeedItem{
					StrDate:       fmt.Sprintf("%d", time.Unix(1, 0).Unix()),
					Description:   "maintenance-desc",
					StrEventStart: fmt.Sprintf("%d", time.Unix(2, 0).Unix()),
					StrEventEnd:   fmt.Sprintf("%d", time.Unix(3, 0).Unix()),
					Title:         "maintenance-title",
					URL:           "https://maintenance.example.com/?entry=1",
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.MaintenanceScheduled,
					metric: createGaugeMetric(1, map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "server",
						"zone":        "is1a",
						"info_url":    "https://maintenance.example.com/?entry=1",
						"info_title":  "maintenance-title",
						"description": "maintenance-desc",
						"start_date":  fmt.Sprintf("%d", time.Unix(2, 0).Unix()),
						"end_date":    fmt.Sprintf("%d", time.Unix(3, 0).Unix()),
					}),
				},
				{
					desc: c.MaintenanceStartTime,
					metric: createGaugeMetric(float64(time.Unix(2, 0).Unix()), map[string]string{
						"id":   "101",
						"name": "server",
						"zone": "is1a",
					}),
				},
				{
					desc: c.MaintenanceEndTime,
					metric: createGaugeMetric(float64(time.Unix(3, 0).Unix()), map[string]string{
						"id":   "101",
						"name": "server",
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

		collected, err := collectMetrics(c, "server")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
