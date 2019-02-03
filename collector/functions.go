package collector

import (
	"fmt"
	"sort"
	"strings"
)

func flattenStringSlice(values []string) string {
	if len(values) == 0 {
		return ""
	}

	sort.Strings(values)
	// NOTE: if values includes comma, this will not work.
	return fmt.Sprintf(",%s,", strings.Join(values, ","))
}
