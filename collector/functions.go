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
	var strValues []string
	for _, v := range values {
		strValues = append(strValues, string(v))
	}
	return flattenStringSlice(strValues)
}
