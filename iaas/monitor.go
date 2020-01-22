// Copyright 2019-2020 The sakuracloud_exporter Authors
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

func monitorLocalRouterValue(values []*sacloud.MonitorLocalRouterValue) *sacloud.MonitorLocalRouterValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}
