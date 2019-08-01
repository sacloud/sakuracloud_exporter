package iaas

import (
	"sort"
	"time"

	"github.com/sacloud/libsacloud/v2/sacloud"
)

func monitorCondition(end time.Time) *sacloud.MonitorCondition {
	end = end.Truncate(time.Second)
	start := end.Add(-time.Hour)
	return &sacloud.MonitorCondition{
		Start: start,
		End:   end,
	}
}

func monitorDatabaseValue(values []*sacloud.MonitorDatabaseValue) *sacloud.MonitorDatabaseValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorCPUTimeValue(values []*sacloud.MonitorCPUTimeValue) *sacloud.MonitorCPUTimeValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorDiskValue(values []*sacloud.MonitorDiskValue) *sacloud.MonitorDiskValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorInterfaceValue(values []*sacloud.MonitorInterfaceValue) *sacloud.MonitorInterfaceValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorRouterValue(values []*sacloud.MonitorRouterValue) *sacloud.MonitorRouterValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorFreeDiskSizeValue(values []*sacloud.MonitorFreeDiskSizeValue) *sacloud.MonitorFreeDiskSizeValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorConnectionValue(values []*sacloud.MonitorConnectionValue) *sacloud.MonitorConnectionValue {
	if len(values) > 2 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorLinkValue(values []*sacloud.MonitorLinkValue) *sacloud.MonitorLinkValue {
	if len(values) > 2 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}
