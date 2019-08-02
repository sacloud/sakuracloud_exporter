package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/v2/sacloud/types"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/sakuracloud_exporter/iaas"
	"github.com/stretchr/testify/require"
)

type dummyAutoBackupClient struct {
	autoBackup     []*sacloud.AutoBackup
	findErr        error
	archives       []*sacloud.Archive
	listBackupsErr error
}

func (d *dummyAutoBackupClient) Find(ctx context.Context) ([]*sacloud.AutoBackup, error) {
	return d.autoBackup, d.findErr
}

func (d *dummyAutoBackupClient) ListBackups(ctx context.Context, zone string, autoBackupID types.ID) ([]*sacloud.Archive, error) {
	return d.archives, d.listBackupsErr
}

func TestAutoBackupCollector_Describe(t *testing.T) {
	initLoggerAndErrors()
	c := NewAutoBackupCollector(context.Background(), testLogger, testErrors, &dummyAutoBackupClient{})

	descs := collectDescs(c)
	require.Len(t, descs, len([]*prometheus.Desc{
		c.Info,
		c.BackupCount,
		c.LastBackupTime,
		c.BackupInfo,
	}))
}

func TestAutoBackupCollector_Collect(t *testing.T) {
	initLoggerAndErrors()
	c := NewAutoBackupCollector(context.Background(), testLogger, testErrors, nil)

	cases := []struct {
		name           string
		in             iaas.AutoBackupClient
		wantLogs       []string
		wantErrCounter float64
		wantMetrics    []*collectedMetric
	}{
		{
			name: "collector returns error",
			in: &dummyAutoBackupClient{
				findErr: errors.New("dummy"),
			},
			wantLogs:       []string{`level=warn msg="can't list autoBackups" err=dummy`},
			wantErrCounter: 1,
			wantMetrics:    nil,
		},
		{
			name:        "empty result",
			in:          &dummyAutoBackupClient{},
			wantMetrics: nil,
		},
		{
			name: "a auto-backup: list archives is failed ",
			in: &dummyAutoBackupClient{
				autoBackup: []*sacloud.AutoBackup{
					{
						ID:                      101,
						Name:                    "AutoBackup",
						DiskID:                  201,
						MaximumNumberOfArchives: 3,
						BackupSpanWeekdays: []types.EBackupSpanWeekday{
							types.BackupSpanWeekdays.Sunday,
							types.BackupSpanWeekdays.Monday,
							types.BackupSpanWeekdays.Tuesday,
						},
						Tags:        types.Tags{"tag1", "tag2"},
						Description: "desc",
					},
				},
				listBackupsErr: errors.New("dummy"),
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Info,
					metric: createGaugeMetric(1, map[string]string{
						"id":             "101",
						"name":           "AutoBackup",
						"disk_id":        "201",
						"max_backup_num": "3",
						"weekdays":       ",sun,mon,tue,",
						"tags":           ",tag1,tag2,",
						"description":    "desc",
					}),
				},
			},
			wantLogs:       []string{`level=warn msg="can't list backed up archives" err=dummy`},
			wantErrCounter: 1,
		},
		{
			name: "a auto-backup without archives",
			in: &dummyAutoBackupClient{
				autoBackup: []*sacloud.AutoBackup{
					{
						ID:                      101,
						Name:                    "AutoBackup",
						DiskID:                  201,
						MaximumNumberOfArchives: 3,
						BackupSpanWeekdays: []types.EBackupSpanWeekday{
							types.BackupSpanWeekdays.Sunday,
							types.BackupSpanWeekdays.Monday,
							types.BackupSpanWeekdays.Tuesday,
						},
						Tags:        types.Tags{"tag1", "tag2"},
						Description: "desc",
					},
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Info,
					metric: createGaugeMetric(1, map[string]string{
						"id":             "101",
						"name":           "AutoBackup",
						"disk_id":        "201",
						"max_backup_num": "3",
						"weekdays":       ",sun,mon,tue,",
						"tags":           ",tag1,tag2,",
						"description":    "desc",
					}),
				},
				{
					desc: c.BackupCount,
					metric: createGaugeMetric(0, map[string]string{
						"id":      "101",
						"name":    "AutoBackup",
						"disk_id": "201",
					}),
				},
				{
					desc: c.LastBackupTime,
					metric: createGaugeMetric(0, map[string]string{
						"id":      "101",
						"name":    "AutoBackup",
						"disk_id": "201",
					}),
				},
			},
		},
		{
			name: "a auto-backup with archives",
			in: &dummyAutoBackupClient{
				autoBackup: []*sacloud.AutoBackup{
					{
						ID:                      101,
						Name:                    "AutoBackup",
						DiskID:                  201,
						MaximumNumberOfArchives: 3,
						BackupSpanWeekdays: []types.EBackupSpanWeekday{
							types.BackupSpanWeekdays.Sunday,
							types.BackupSpanWeekdays.Monday,
							types.BackupSpanWeekdays.Tuesday,
						},
						Tags:        types.Tags{"tag1", "tag2"},
						Description: "desc",
					},
				},
				archives: []*sacloud.Archive{
					{
						ID:          301,
						Name:        "Archive1",
						Tags:        types.Tags{"tag1-1", "tag1-2"},
						Description: "desc1",
						CreatedAt:   time.Unix(1, 0),
					},
					{
						ID:          302,
						Name:        "Archive2",
						Tags:        types.Tags{"tag2-1", "tag2-2"},
						Description: "desc2",
						CreatedAt:   time.Unix(2, 0),
					},
				},
			},
			wantMetrics: []*collectedMetric{
				{
					desc: c.Info,
					metric: createGaugeMetric(1, map[string]string{
						"id":             "101",
						"name":           "AutoBackup",
						"disk_id":        "201",
						"max_backup_num": "3",
						"weekdays":       ",sun,mon,tue,",
						"tags":           ",tag1,tag2,",
						"description":    "desc",
					}),
				},
				{
					// BackupCount
					desc: c.BackupCount,
					metric: createGaugeMetric(2, map[string]string{
						"id":      "101",
						"name":    "AutoBackup",
						"disk_id": "201",
					}),
				},
				{
					// LastBackupTime: latest backup is created at time.Unix(2,0), so value is 2000(milli sec)
					desc: c.LastBackupTime,
					metric: createGaugeMetric(2000, map[string]string{
						"id":      "101",
						"name":    "AutoBackup",
						"disk_id": "201",
					}),
				},
				{
					// backup1
					desc: c.BackupInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                  "101",
						"name":                "AutoBackup",
						"disk_id":             "201",
						"archive_id":          "301",
						"archive_name":        "Archive1",
						"archive_description": "desc1",
						"archive_tags":        ",tag1-1,tag1-2,",
					}),
				},
				{
					// backup2
					desc: c.BackupInfo,
					metric: createGaugeMetric(1, map[string]string{
						"id":                  "101",
						"name":                "AutoBackup",
						"disk_id":             "201",
						"archive_id":          "302",
						"archive_name":        "Archive2",
						"archive_description": "desc2",
						"archive_tags":        ",tag2-1,tag2-2,",
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

		collected, err := collectMetrics(c, "auto_backup")
		require.NoError(t, err)
		require.Equal(t, tc.wantLogs, collected.logged)
		require.Equal(t, tc.wantErrCounter, *collected.errors.Counter.Value)
		requireMetricsEqual(t, tc.wantMetrics, collected.collected)
	}
}
