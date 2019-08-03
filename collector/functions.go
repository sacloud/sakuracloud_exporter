package collector

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

func flattenStringSlice(values []string) string {
	if len(values) == 0 {
		return ""
	}

	sort.Strings(values)
	// NOTE: if values includes comma, this will not work.
	return fmt.Sprintf(",%s,", strings.Join(values, ","))
}

func flattenBackupSpanWeekdays(values []types.EBackupSpanWeekday) string {
	if len(values) == 0 {
		return ""
	}

	// sort
	sort.Slice(values, func(i, j int) bool {
		return backupSpanWeekdaysOrder[values[i]] < backupSpanWeekdaysOrder[values[j]]
	})

	var strValues []string
	for _, v := range values {
		strValues = append(strValues, string(v))
	}
	return fmt.Sprintf(",%s,", strings.Join(strValues, ","))
}

var backupSpanWeekdaysOrder = map[types.EBackupSpanWeekday]int{
	types.BackupSpanWeekdays.Sunday:    0,
	types.BackupSpanWeekdays.Monday:    1,
	types.BackupSpanWeekdays.Tuesday:   2,
	types.BackupSpanWeekdays.Wednesday: 3,
	types.BackupSpanWeekdays.Thursday:  4,
	types.BackupSpanWeekdays.Friday:    5,
	types.BackupSpanWeekdays.Saturday:  6,
}
