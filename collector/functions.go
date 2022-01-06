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
