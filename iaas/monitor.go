package iaas

import (
	"sort"
	"time"

	sakuraAPI "github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
)

// NICMetrics represents NIC's receive/send metrics
type NICMetrics struct {
	Receive *sacloud.FlatMonitorValue
	Send    *sacloud.FlatMonitorValue
}

// RouterMetrics represents Switch+Router(Internet)'s in/out traffic metrics
type RouterMetrics struct {
	In  *sacloud.FlatMonitorValue
	Out *sacloud.FlatMonitorValue
}

// SIMMetrics represents SIM uplink/downlink metrics
type SIMMetrics struct {
	Uplink   *sacloud.FlatMonitorValue
	Downlink *sacloud.FlatMonitorValue
}

// DiskMetrics represents Disk's read/write metrics
type DiskMetrics struct {
	Read  *sacloud.FlatMonitorValue
	Write *sacloud.FlatMonitorValue
}

// DatabaseMetrics represents Database's system metrics
type DatabaseMetrics struct {
	TotalMemorySize   *sacloud.FlatMonitorValue
	UsedMemorySize    *sacloud.FlatMonitorValue
	TotalDisk1Size    *sacloud.FlatMonitorValue
	UsedDisk1Size     *sacloud.FlatMonitorValue
	TotalDisk2Size    *sacloud.FlatMonitorValue
	UsedDisk2Size     *sacloud.FlatMonitorValue
	DelayTimeSec      *sacloud.FlatMonitorValue
	BinlogUsedSizeKiB *sacloud.FlatMonitorValue
}

type queryMonitorFn func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error)

func queryCPUTimeMonitorValue(
	client *sakuraAPI.Client,
	zone string, end time.Time,
	queryFn queryMonitorFn) (*sacloud.FlatMonitorValue, error) {

	mv, err := queryMonitorValues(client, zone, end, queryFn)
	if err != nil {
		return nil, err
	}
	if mv == nil {
		return nil, nil
	}

	// find latest value
	values, err := mv.FlattenCPUTimeValue()
	if err != nil {
		return nil, err
	}

	return getLatestMonitorValue(values), nil
}

func queryFreeDiskSizeMonitorValue(
	client *sakuraAPI.Client,
	zone string, end time.Time,
	queryFn queryMonitorFn) (*sacloud.FlatMonitorValue, error) {

	mv, err := queryMonitorValues(client, zone, end, queryFn)
	if err != nil {
		return nil, err
	}
	if mv == nil {
		return nil, nil
	}

	// find latest value
	values, err := mv.FlattenFreeDiskSizeValue()
	if err != nil {
		return nil, err
	}

	return getLatestMonitorValue(values), nil
}

func queryNICMonitorValue(
	client *sakuraAPI.Client,
	zone string, end time.Time,
	queryFn queryMonitorFn) (*NICMetrics, error) {

	mv, err := queryMonitorValues(client, zone, end, queryFn)
	if err != nil {
		return nil, err
	}
	if mv == nil {
		return nil, nil
	}

	receive, err := mv.FlattenPacketReceiveValue()
	if err != nil {
		return nil, err
	}
	send, err := mv.FlattenPacketSendValue()
	if err != nil {
		return nil, err
	}

	metrics := &NICMetrics{
		Receive: getLatestMonitorValue(receive),
		Send:    getLatestMonitorValue(send),
	}
	if metrics.Receive == nil || metrics.Send == nil {
		return nil, nil
	}

	return metrics, nil
}

func queryRouterMonitorValue(
	client *sakuraAPI.Client,
	zone string, end time.Time,
	queryFn queryMonitorFn) (*RouterMetrics, error) {

	mv, err := queryMonitorValues(client, zone, end, queryFn)
	if err != nil {
		return nil, err
	}
	if mv == nil {
		return nil, nil
	}

	in, err := mv.FlattenInternetInValue()
	if err != nil {
		return nil, err
	}
	out, err := mv.FlattenInternetOutValue()
	if err != nil {
		return nil, err
	}

	metrics := &RouterMetrics{
		In:  getLatestMonitorValue(in),
		Out: getLatestMonitorValue(out),
	}
	if metrics.In == nil || metrics.Out == nil {
		return nil, nil
	}

	return metrics, nil
}

func querySIMMonitorValue(
	client *sakuraAPI.Client,
	zone string, end time.Time,
	queryFn queryMonitorFn) (*SIMMetrics, error) {

	mv, err := queryMonitorValues(client, zone, end, queryFn)
	if err != nil {
		return nil, err
	}
	if mv == nil {
		return nil, nil
	}

	uplink, err := mv.FlattenUplinkBPSValue()
	if err != nil {
		return nil, err
	}
	downlink, err := mv.FlattenDownlinkBPSValue()
	if err != nil {
		return nil, err
	}

	metrics := &SIMMetrics{
		Uplink:   getLatestMonitorValue(uplink),
		Downlink: getLatestMonitorValue(downlink),
	}
	if metrics.Uplink == nil || metrics.Downlink == nil {
		return nil, nil
	}

	return metrics, nil
}

func queryDiskMonitorValue(
	client *sakuraAPI.Client,
	zone string, end time.Time,
	queryFn queryMonitorFn) (*DiskMetrics, error) {

	mv, err := queryMonitorValues(client, zone, end, queryFn)
	if err != nil {
		return nil, err
	}
	if mv == nil {
		return nil, nil
	}

	read, err := mv.FlattenDiskReadValue()
	if err != nil {
		return nil, err
	}
	write, err := mv.FlattenDiskWriteValue()
	if err != nil {
		return nil, err
	}

	metrics := &DiskMetrics{
		Read:  getLatestMonitorValue(read),
		Write: getLatestMonitorValue(write),
	}
	if metrics.Read == nil || metrics.Write == nil {
		return nil, nil
	}

	return metrics, nil
}

func queryDatabaseMonitorValue(
	client *sakuraAPI.Client,
	zone string, end time.Time,
	queryFn queryMonitorFn) (*DatabaseMetrics, error) {

	mv, err := queryMonitorValues(client, zone, end, queryFn)
	if err != nil {
		return nil, err
	}
	if mv == nil {
		return nil, nil
	}

	totalMemory, err := mv.FlattenTotalMemorySizeValue()
	if err != nil {
		return nil, err
	}
	usedMemory, err := mv.FlattenUsedMemorySizeValue()
	if err != nil {
		return nil, err
	}
	totalDisk1Size, err := mv.FlattenTotalDisk1SizeValue()
	if err != nil {
		return nil, err
	}
	usedDisk1Size, err := mv.FlattenUsedDisk1SizeValue()
	if err != nil {
		return nil, err
	}
	totalDisk2Size, err := mv.FlattenTotalDisk2SizeValue()
	if err != nil {
		return nil, err
	}
	usedDisk2Size, err := mv.FlattenUsedDisk2SizeValue()
	if err != nil {
		return nil, err
	}
	delayTime, err := mv.FlattenDelayTimeSecValue()
	if err != nil {
		return nil, err
	}
	binlogSize, err := mv.FlattenBinlogUsedSizeKiBValue()
	if err != nil {
		return nil, err
	}

	metrics := &DatabaseMetrics{
		TotalMemorySize:   getLatestMonitorValue(totalMemory),
		UsedMemorySize:    getLatestMonitorValue(usedMemory),
		TotalDisk1Size:    getLatestMonitorValue(totalDisk1Size),
		UsedDisk1Size:     getLatestMonitorValue(usedDisk1Size),
		TotalDisk2Size:    getLatestMonitorValue(totalDisk2Size),
		UsedDisk2Size:     getLatestMonitorValue(usedDisk2Size),
		DelayTimeSec:      getLatestMonitorValue(delayTime),
		BinlogUsedSizeKiB: getLatestMonitorValue(binlogSize),
	}

	if metrics.TotalMemorySize == nil &&
		metrics.UsedMemorySize == nil &&
		metrics.TotalDisk1Size == nil &&
		metrics.UsedDisk1Size == nil &&
		metrics.TotalDisk2Size == nil &&
		metrics.UsedDisk2Size == nil &&
		metrics.DelayTimeSec == nil &&
		metrics.BinlogUsedSizeKiB == nil {
		return nil, nil
	}
	return metrics, nil
}

func queryMonitorValues(
	client *sakuraAPI.Client,
	zone string, end time.Time,
	queryFn queryMonitorFn) (*sacloud.MonitorValues, error) {

	c := client.Clone()
	c.Zone = zone

	start := end.Add(-30 * time.Minute)
	param := sacloud.NewResourceMonitorRequest(&start, &end)

	return queryFn(c, param)
}

func getLatestMonitorValue(values []sacloud.FlatMonitorValue) *sacloud.FlatMonitorValue {

	// Note: Latest value is temporary(this is API spec). so use Latest+1 value
	if len(values) < 2 {
		return nil
	}

	// Descending
	sort.Slice(values, func(i, j int) bool { return values[i].Time.After(values[j].Time) })

	return &values[1]
}
