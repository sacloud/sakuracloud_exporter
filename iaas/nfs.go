package iaas

import (
	"time"

	sakuraAPI "github.com/sacloud/libsacloud/api"

	"github.com/sacloud/libsacloud/sacloud"
)

type NFS struct {
	*sacloud.NFS
	ZoneName string
}

type NFSClient interface {
	Find() ([]*NFS, error)
	MonitorFreeDiskSize(zone string, nfsID int64, end time.Time) (*sacloud.FlatMonitorValue, error)
	MonitorNIC(zone string, nfsID int64, end time.Time) (*NICMetrics, error)
}

func getNFSClient(client *sakuraAPI.Client, zones []string) NFSClient {
	return &nfsClient{
		rawClient: client,
		zones:     zones,
	}
}

type nfsClient struct {
	rawClient *sakuraAPI.Client
	zones     []string
}

func (s *nfsClient) find(c *sakuraAPI.Client) ([]interface{}, error) {
	var results []interface{}
	res, err := c.NFS.Reset().Limit(10000).Find()
	if err != nil {
		return results, err
	}
	for i := range res.NFS {
		results = append(results, &NFS{
			NFS:      &res.NFS[i],
			ZoneName: c.Zone,
		})
	}
	return results, err
}

func (s *nfsClient) Find() ([]*NFS, error) {
	res, err := queryToZones(s.rawClient, s.zones, s.find)
	if err != nil {
		return nil, err
	}
	var results []*NFS
	for _, s := range res {
		results = append(results, s.(*NFS))
	}
	return results, nil
}

func (s *nfsClient) MonitorFreeDiskSize(zone string, nfsID int64, end time.Time) (*sacloud.FlatMonitorValue, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.NFS.MonitorFreeDiskSize(nfsID, param)
	}

	return queryFreeDiskSizeMonitorValue(s.rawClient, zone, end, query)
}

func (s *nfsClient) MonitorNIC(zone string, nfsID int64, end time.Time) (*NICMetrics, error) {
	query := func(client *sakuraAPI.Client, param *sacloud.ResourceMonitorRequest) (*sacloud.MonitorValues, error) {
		return client.NFS.MonitorInterface(nfsID, param)
	}

	return queryNICMonitorValue(s.rawClient, zone, end, query)
}
