// Copyright 2019-2025 The sakuracloud_exporter Authors
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

package platform

import (
	"sort"
	"time"

	"github.com/sacloud/iaas-api-go"
)

func monitorCondition(end time.Time) *iaas.MonitorCondition {
	end = end.Truncate(time.Second)
	start := end.Add(-time.Hour)
	return &iaas.MonitorCondition{
		Start: start,
		End:   end,
	}
}

func monitorDatabaseValue(values []*iaas.MonitorDatabaseValue) *iaas.MonitorDatabaseValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorCPUTimeValue(values []*iaas.MonitorCPUTimeValue) *iaas.MonitorCPUTimeValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorDiskValue(values []*iaas.MonitorDiskValue) *iaas.MonitorDiskValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorInterfaceValue(values []*iaas.MonitorInterfaceValue) *iaas.MonitorInterfaceValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorRouterValue(values []*iaas.MonitorRouterValue) *iaas.MonitorRouterValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorFreeDiskSizeValue(values []*iaas.MonitorFreeDiskSizeValue) *iaas.MonitorFreeDiskSizeValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorConnectionValue(values []*iaas.MonitorConnectionValue) *iaas.MonitorConnectionValue {
	if len(values) > 2 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorLinkValue(values []*iaas.MonitorLinkValue) *iaas.MonitorLinkValue {
	if len(values) > 2 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}

func monitorLocalRouterValue(values []*iaas.MonitorLocalRouterValue) *iaas.MonitorLocalRouterValue {
	if len(values) > 1 {
		// Descending
		sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })
		return values[1]
	}
	return nil
}
