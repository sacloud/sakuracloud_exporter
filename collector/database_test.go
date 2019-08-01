package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/sacloud/libsacloud/v2/sacloud/types"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/sakuracloud_exporter/iaas"
	"github.com/stretchr/testify/require"
)

type dummyDatabaseClient struct {
	find           []*iaas.Database
	findErr        error
	monitorDB      *sacloud.MonitorDatabaseValue
	monitorDBErr   error
	monitorCPU     *sacloud.MonitorCPUTimeValue
	monitorCPUErr  error
	monitorNIC     *sacloud.MonitorInterfaceValue
	monitorNICErr  error
	monitorDisk    *sacloud.MonitorDiskValue
	monitorDiskErr error
}

func (d *dummyDatabaseClient) Find(ctx context.Context) ([]*iaas.Database, error) {
	return d.find, d.findErr
}
func (d *dummyDatabaseClient) MonitorDatabase(ctx context.Context, zone string, diskID types.ID, end time.Time) (*sacloud.MonitorDatabaseValue, error) {
	return d.monitorDB, d.monitorDBErr
}
func (d *dummyDatabaseClient) MonitorCPU(ctx context.Context, zone string, databaseID types.ID, end time.Time) (*sacloud.MonitorCPUTimeValue, error) {
	return d.monitorCPU, d.monitorCPUErr
}
func (d *dummyDatabaseClient) MonitorNIC(ctx context.Context, zone string, databaseID types.ID, end time.Time) (*sacloud.MonitorInterfaceValue, error) {
	return d.monitorNIC, d.monitorNICErr
}
func (d *dummyDatabaseClient) MonitorDisk(ctx context.Context, zone string, databaseID types.ID, end time.Time) (*sacloud.MonitorDiskValue, error) {
	return d.monitorDisk, d.monitorDiskErr
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
	}))
}

func TestDatabaseCollector_Collect(t *testing.T) {
	cases := []struct {
		name           string
		in             iaas.DatabaseClient
		wantLog        string
		wantErrCounter float64
		wantMetrics    []*dto.Metric
	}{
		{
			name: "collector returns error",
			in: &dummyDatabaseClient{
				findErr: errors.New("dummy"),
			},
			wantLog:        `level=warn msg="can't list databases" err=dummy`,
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:           "empty result",
			in:             &dummyDatabaseClient{},
			wantLog:        "",
			wantErrCounter: 0,
			wantMetrics:    nil,
		},
	}

	for _, tc := range cases {
		initLoggerAndErrors()
		c := NewDatabaseCollector(context.Background(), testLogger, testErrors, tc.in)
		collected, err := collectMetrics(c, "database")
		require.NoError(t, err)
		require.Equal(t, tc.wantLog, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		require.Equal(t, tc.wantMetrics, collected.collected)
	}
}
