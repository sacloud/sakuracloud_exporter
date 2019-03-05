package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/sacloud/sakuracloud_exporter/iaas"
)

// NFSCollector collects metrics about all nfss.
type NFSCollector struct {
	logger log.Logger
	errors *prometheus.CounterVec
	client iaas.NFSClient

	Up      *prometheus.Desc
	NFSInfo *prometheus.Desc

	DiskFree *prometheus.Desc

	NICInfo    *prometheus.Desc
	NICReceive *prometheus.Desc
	NICSend    *prometheus.Desc
}

// NewNFSCollector returns a new NFSCollector.
func NewNFSCollector(logger log.Logger, errors *prometheus.CounterVec, client iaas.NFSClient) *NFSCollector {
	errors.WithLabelValues("nfs").Add(0)

	nfsLabels := []string{"id", "name", "zone"}
	nfsInfoLabels := append(nfsLabels, "plan", "host", "tags", "description")
	nicInfoLabels := append(nfsLabels, "upstream_id", "upstream_name", "ipaddress", "nw_mask_len", "gateway")

	return &NFSCollector{
		logger: logger,
		errors: errors,
		client: client,
		Up: prometheus.NewDesc(
			"sakuracloud_nfs_up",
			"If 1 the nfs is up and running, 0 otherwise",
			nfsLabels, nil,
		),
		NFSInfo: prometheus.NewDesc(
			"sakuracloud_nfs_info",
			"A metric with a constant '1' value labeled by nfs information",
			nfsInfoLabels, nil,
		),
		DiskFree: prometheus.NewDesc(
			"sakuracloud_nfs_free_disk_size",
			"NFS's Free Disk Size(unit: GB)",
			nfsLabels, nil,
		),
		NICInfo: prometheus.NewDesc(
			"sakuracloud_nfs_nic_info",
			"A metric with a constant '1' value labeled by nic information",
			nicInfoLabels, nil,
		),
		NICReceive: prometheus.NewDesc(
			"sakuracloud_nfs_receive",
			"NIC's receive bytes(unit: Kbps)",
			nfsLabels, nil,
		),
		NICSend: prometheus.NewDesc(
			"sakuracloud_nfs_send",
			"NIC's send bytes(unit: Kbps)",
			nfsLabels, nil,
		),
	}
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector.
func (c *NFSCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Up
	ch <- c.NFSInfo
	ch <- c.DiskFree
	ch <- c.NICInfo
	ch <- c.NICReceive
	ch <- c.NICSend
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *NFSCollector) Collect(ch chan<- prometheus.Metric) {
	nfss, err := c.client.Find()
	if err != nil {
		c.errors.WithLabelValues("nfs").Add(1)
		level.Warn(c.logger).Log(
			"msg", "can't list nfss",
			"err", err,
		)
	}

	var wg sync.WaitGroup
	wg.Add(len(nfss))

	for i := range nfss {
		go func(nfs *iaas.NFS) {
			defer wg.Done()

			nfsLabels := c.nfsLabels(nfs)

			var up float64
			if nfs.IsUp() {
				up = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Up,
				prometheus.GaugeValue,
				up,
				nfsLabels...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.NFSInfo,
				prometheus.GaugeValue,
				float64(1.0),
				c.nfsInfoLabels(nfs)...,
			)
			ch <- prometheus.MustNewConstMetric(
				c.NICInfo,
				prometheus.GaugeValue,
				float64(1.0),
				c.nicInfoLabels(nfs)...,
			)

			if nfs.IsUp() {
				now := time.Now()
				// Free disk size
				wg.Add(1)
				go func() {
					c.collectFreeDiskSize(ch, nfs, now)
					wg.Done()
				}()

				// NICs
				wg.Add(1)
				go func() {
					c.collectNICMetrics(ch, nfs, now)
					wg.Done()
				}()
			}

		}(nfss[i])
	}

	wg.Wait()
}

func (c *NFSCollector) nfsLabels(nfs *iaas.NFS) []string {
	return []string{
		nfs.GetStrID(),
		nfs.Name,
		nfs.ZoneName,
	}
}

var nfsPlanLabels = map[int64]string{
	int64(sacloud.NFSPlan100G): "100GB",
	int64(sacloud.NFSPlan500G): "500GB",
	int64(sacloud.NFSPlan1T):   "1TB",
	int64(sacloud.NFSPlan2T):   "2TB",
	int64(sacloud.NFSPlan4T):   "4TB",
}

func (c *NFSCollector) nfsInfoLabels(nfs *iaas.NFS) []string {
	labels := c.nfsLabels(nfs)

	instanceHost := "-"
	if nfs.Instance != nil {
		instanceHost = nfs.Instance.Host.Name
	}

	return append(labels,
		nfsPlanLabels[nfs.Plan.ID],
		instanceHost,
		flattenStringSlice(nfs.Tags),
		nfs.Description,
	)
}

func (c *NFSCollector) nicInfoLabels(nfs *iaas.NFS) []string {
	labels := c.nfsLabels(nfs)

	upstreamID := nfs.Switch.GetStrID()
	upstreamName := nfs.Switch.Name

	nwMaskLen := nfs.NetworkMaskLen()
	strMaskLen := ""
	if nwMaskLen > 0 {
		strMaskLen = fmt.Sprintf("%d", nwMaskLen)
	}

	return append(labels,
		upstreamID,
		upstreamName,
		nfs.IPAddress(),
		strMaskLen,
		nfs.DefaultRoute(),
	)
}

func (c *NFSCollector) collectFreeDiskSize(ch chan<- prometheus.Metric, nfs *iaas.NFS, now time.Time) {

	values, err := c.client.MonitorFreeDiskSize(nfs.ZoneName, nfs.ID, now)
	if err != nil {
		c.errors.WithLabelValues("nfs").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get disk's free size: NFSID=%d", nfs.ID),
			"err", err,
		)
		return
	}
	if len(values) == 0 {
		return
	}

	for _, v := range values {
		if v.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.DiskFree,
				prometheus.GaugeValue,
				v.Value/1024/1024, // unit:GB
				c.nfsLabels(nfs)...,
			)

			ch <- prometheus.NewMetricWithTimestamp(v.Time, m)
		}
	}
}

func (c *NFSCollector) collectNICMetrics(ch chan<- prometheus.Metric, nfs *iaas.NFS, now time.Time) {

	values, err := c.client.MonitorNIC(nfs.ZoneName, nfs.ID, now)
	if err != nil {
		c.errors.WithLabelValues("nfs").Add(1)
		level.Warn(c.logger).Log(
			"msg", fmt.Sprintf("can't get nfs's NIC metrics: NFSID=%d", nfs.ID),
			"err", err,
		)
		return
	}
	if len(values) == 0 {
		return
	}

	for _, v := range values {
		if v.Receive != nil && v.Receive.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.NICReceive,
				prometheus.GaugeValue,
				v.Receive.Value*8/1000,
				c.nfsLabels(nfs)...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.Receive.Time, m)
		}
		if v.Send != nil && v.Send.Time.Unix() > 0 {
			m := prometheus.MustNewConstMetric(
				c.NICSend,
				prometheus.GaugeValue,
				v.Send.Value*8/1000,
				c.nfsLabels(nfs)...,
			)
			ch <- prometheus.NewMetricWithTimestamp(v.Send.Time, m)
		}
	}
}
