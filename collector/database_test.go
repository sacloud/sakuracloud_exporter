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
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/newsfeed"
	"github.com/sacloud/sakuracloud_exporter/platform"
	"github.com/stretchr/testify/require"
)

type dummyDatabaseClient struct {
	find           []*platform.Database
	findErr        error
	monitorDB      *iaas.MonitorDatabaseValue
	monitorDBErr   error
	monitorCPU     *iaas.MonitorCPUTimeValue
	monitorCPUErr  error
	monitorNIC     *iaas.MonitorInterfaceValue
	monitorNICErr  error
	monitorDisk    *iaas.MonitorDiskValue
	monitorDiskErr error
	maintenance    *newsfeed.FeedItem
	maintenanceErr error
}

func (d *dummyDatabaseClient) Find(ctx context.Context) ([]*platform.Database, error) {
	return d.find, d.findErr
}
func (d *dummyDatabaseClient) MonitorDatabase(ctx context.Context, zone string, diskID types.ID, end time.Time) (*iaas.MonitorDatabaseValue, error) {
	return d.monitorDB, d.monitorDBErr
}
func (d *dummyDatabaseClient) MonitorCPU(ctx context.Context, zone string, databaseID types.ID, end time.Time) (*iaas.MonitorCPUTimeValue, error) {
	return d.monitorCPU, d.monitorCPUErr
}
func (d *dummyDatabaseClient) MonitorNIC(ctx context.Context, zone string, databaseID types.ID, end time.Time) (*iaas.MonitorInterfaceValue, error) {
	return d.monitorNIC, d.monitorNICErr
}
func (d *dummyDatabaseClient) MonitorDisk(ctx context.Context, zone string, databaseID types.ID, end time.Time) (*iaas.MonitorDiskValue, error) {
	return d.monitorDisk, d.monitorDiskErr
}
func (d *dummyDatabaseClient) MaintenanceInfo(infoURL string) (*newsfeed.FeedItem, error) {
	return d.maintenance, d.maintenanceErr
}

func TestDatabaseCollector_Describe(t *testing.T) {
	initLoggerAndErrors()

	c := NewDatabaseCollector(context.Background(), testLogger, testErrors, &dummyDatabaseClient{})
	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Up,
		c.DatabaseInfo,
		c.CPUTime,
		c.MemoryUsed,
		c.MemoryTotal,
		c.NICInfo,
		c.NICReceive,
		c.NICSend,
		c.SystemDiskUsed,
		c.SystemDiskTotal,
		c.BackupDiskUsed,
		c.BackupDiskTotal,
		c.BinlogUsed,
		c.DiskRead,
		c.DiskWrite,
		c.ReplicationDelay,
		c.MaintenanceScheduled,
		c.MaintenanceInfo,
		c.MaintenanceStartTime,
		c.MaintenanceEndTime,
	}))
}

func TestDatabaseCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewDatabaseCollector(context.Background(), testLogger, testErrors, nil)

	var (
		dbValue = &platform.Database{
			Database: &iaas.Database{
				ID:               101,
				Name:             "database",
				Description:      "desc",
				Tags:             types.Tags{"tag1", "tag2"},
				Availability:     types.Availabilities.Available,
				InstanceStatus:   types.ServerInstanceStatuses.Up,
				InstanceHostName: "sacXXXX",
				PlanID:           types.DatabasePlans.DB10GB,
				Conf: &iaas.DatabaseRemarkDBConfCommon{
					DatabaseName:     types.RDBMSTypesMariaDB.String(),
					DatabaseVersion:  "1",
					DatabaseRevision: "1",
				},
				Interfaces: []*iaas.InterfaceView{
					{
						ID:           201,
						UpstreamType: types.UpstreamNetworkTypes.Switch,
						SwitchID:     301,
						SwitchName:   "switch",
					},
				},
				IPAddresses:    []string{"192.168.0.11"},
				NetworkMaskLen: 24,
				DefaultRoute:   "192.168.0.1",
			},
			ZoneName: "is1a",
		}
		dbLabels = map[string]string{
			"id":   "101",
			"name": "database",
			"zone": "is1a",
		}
		dbInfoLabels = map[string]string{
			"id":                  "101",
			"name":                "database",
			"zone":                "is1a",
			"plan":                "10GB",
			"host":                "sacXXXX",
			"database_type":       types.RDBMSTypesMariaDB.String(),
			"database_revision":   "1",
			"database_version":    "1",
			"web_ui":              "",
			"replication_enabled": "0",
			"replication_role":    "",
			"tags":                ",tag1,tag2,",
			"description":         "desc",
		}
		nicInfoLabels = map[string]string{
			"id":            "101",
			"name":          "database",
			"zone":          "is1a",
			"upstream_type": "switch",
			"upstream_id":   "301",
			"upstream_name": "switch",
			"ipaddress":     "192.168.0.11",
			"nw_mask_len":   "24",
			"gateway":       "192.168.0.1",
		}
		monitorTime = time.Unix(1, 0)
	)

	cases := []struct {
		name           string
		in             platform.DatabaseClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyDatabaseClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=WARN msg="can't list databases" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyDatabaseClient{},
			wantMetrics: nil,
		},
		{
			name: "a database without activity monitor values",
			in: &dummyDatabaseClient{
				find: []*platform.Database{
					dbValue,
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc:   c.Up,
					metric: createGaugeMetric(1, dbLabels),
				},
				{
					desc:   c.DatabaseInfo,
					metric: createGaugeMetric(1, dbInfoLabels),
				},
				{
					desc:   c.NICInfo,
					metric: createGaugeMetric(1, nicInfoLabels),
				},
				{
					desc:   c.MaintenanceScheduled,
					metric: createGaugeMetric(0, dbLabels),
				},
			},
		},
		{
			name: "activity monitor returns error",
			in: &dummyDatabaseClient{
				find: []*platform.Database{
					dbValue,
				},
				monitorDBErr:   errors.New("dummy"),
				monitorCPUErr:  errors.New("dummy"),
				monitorNICErr:  errors.New("dummy"),
				monitorDiskErr: errors.New("dummy"),
			},
			wantMetrics: []*collectedMetric{
				{
					desc:   c.Up,
					metric: createGaugeMetric(1, dbLabels),
				},
				{
					desc:   c.DatabaseInfo,
					metric: createGaugeMetric(1, dbInfoLabels),
				},
				{
					desc:   c.NICInfo,
					metric: createGaugeMetric(1, nicInfoLabels),
				},
				{
					desc:   c.MaintenanceScheduled,
					metric: createGaugeMetric(0, dbLabels),
				},
			},
			wantLogs: []string{
				`level=WARN msg="can't get database's NIC metrics: DatabaseID=101" err=dummy`,
				`level=WARN msg="can't get database's cpu time: DatabaseID=101" err=dummy`,
				`level=WARN msg="can't get database's system metrics: DatabaseID=101" err=dummy`,
				`level=WARN msg="can't get disk's metrics: DatabaseID=101" err=dummy`,
			},
			wantErrCounter: 4,
		},
		{
			name: "all metrics without errors",
			in: &dummyDatabaseClient{
				find: []*platform.Database{
					{
						Database: &iaas.Database{
							ID:           101,
							Name:         "database",
							Description:  "desc",
							Tags:         types.Tags{"tag1", "tag2"},
							Availability: types.Availabilities.Available,

							InstanceStatus:      types.ServerInstanceStatuses.Up,
							InstanceHostName:    "sacXXXX",
							InstanceHostInfoURL: "http://example.com/maintenance-info-dummy-url",

							PlanID: types.DatabasePlans.DB10GB,
							Conf: &iaas.DatabaseRemarkDBConfCommon{
								DatabaseName:     types.RDBMSTypesMariaDB.String(),
								DatabaseVersion:  "1",
								DatabaseRevision: "1",
							},
							Interfaces: []*iaas.InterfaceView{
								{
									ID:           201,
									UpstreamType: types.UpstreamNetworkTypes.Switch,
									SwitchID:     301,
									SwitchName:   "switch",
								},
							},
							IPAddresses:    []string{"192.168.0.11"},
							NetworkMaskLen: 24,
							DefaultRoute:   "192.168.0.1",
						},
						ZoneName: "is1a",
					},
				},
				monitorCPU: &iaas.MonitorCPUTimeValue{
					Time:    monitorTime,
					CPUTime: 101,
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
				monitorDB: &iaas.MonitorDatabaseValue{
					Time:              monitorTime,
					UsedMemorySize:    401,
					TotalMemorySize:   402,
					UsedDisk1Size:     403,
					TotalDisk1Size:    404,
					UsedDisk2Size:     405,
					TotalDisk2Size:    406,
					BinlogUsedSizeKiB: 407,
					DelayTimeSec:      408,
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
					desc:   c.Up,
					metric: createGaugeMetric(1, dbLabels),
				},
				{
					desc:   c.DatabaseInfo,
					metric: createGaugeMetric(1, dbInfoLabels),
				},
				{
					desc:   c.NICInfo,
					metric: createGaugeMetric(1, nicInfoLabels),
				},
				{
					desc:   c.CPUTime,
					metric: createGaugeWithTimestamp(101*1000, dbLabels, monitorTime),
				},
				{
					desc:   c.DiskRead,
					metric: createGaugeWithTimestamp(float64(201)/1024, dbLabels, monitorTime),
				},
				{
					desc:   c.DiskWrite,
					metric: createGaugeWithTimestamp(float64(202)/1024, dbLabels, monitorTime),
				},
				{
					desc:   c.NICReceive,
					metric: createGaugeWithTimestamp(float64(301)*8/1000, dbLabels, monitorTime),
				},
				{
					desc:   c.NICSend,
					metric: createGaugeWithTimestamp(float64(302)*8/1000, dbLabels, monitorTime),
				},
				{
					desc:   c.MemoryUsed,
					metric: createGaugeWithTimestamp(float64(401)/1024/1024, dbLabels, monitorTime),
				},
				{
					desc:   c.MemoryTotal,
					metric: createGaugeWithTimestamp(float64(402)/1024/1024, dbLabels, monitorTime),
				},
				{
					desc:   c.SystemDiskUsed,
					metric: createGaugeWithTimestamp(float64(403)/1024/1024, dbLabels, monitorTime),
				},
				{
					desc:   c.SystemDiskTotal,
					metric: createGaugeWithTimestamp(float64(404)/1024/1024, dbLabels, monitorTime),
				},
				{
					desc:   c.BackupDiskUsed,
					metric: createGaugeWithTimestamp(float64(405)/1024/1024, dbLabels, monitorTime),
				},
				{
					desc:   c.BackupDiskTotal,
					metric: createGaugeWithTimestamp(float64(406)/1024/1024, dbLabels, monitorTime),
				},
				{
					desc:   c.BinlogUsed,
					metric: createGaugeWithTimestamp(float64(407)/1024/1024, dbLabels, monitorTime),
				},
				{
					desc:   c.ReplicationDelay,
					metric: createGaugeWithTimestamp(float64(408), dbLabels, monitorTime),
				},
				{
					desc:   c.MaintenanceScheduled,
					metric: createGaugeMetric(1, dbLabels),
				},
				{
					desc: c.MaintenanceInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":          "101",
						"name":        "database",
						"zone":        "is1a",
						"info_url":    "http://example.com/maintenance",
						"info_title":  "dummy-title",
						"description": "desc",
						"start_date":  "946652400",
						"end_date":    "949244400",
					}),
				},
				{
					desc:   c.MaintenanceStartTime,
					metric: createGaugeMetric(946652400, dbLabels),
				},
				{
					desc:   c.MaintenanceEndTime,
					metric: createGaugeMetric(949244400, dbLabels),
				},
			},
		},
	}

	for _, tc := range cases {
		initLoggerAndErrors()
		c.logger = testLogger
		c.errors = testErrors
		c.client = tc.in

		collected, err := collectMetrics(c, "database")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
